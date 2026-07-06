# True Ambient Capture — plan

> Status: plan for discussion, nothing built. Follows the shipped v0.1.3 (git-hook
> capture + `doctor`). Goal: make capture *can't-miss* and make expansion *ambient*,
> without coupling the two — and decide how hreysi plugs into Daniel Miessler's PAI
> / LifeOS or openclaw. Research-backed; sources at the end.

---

## 0. The problem, split correctly

"Ambient" is two different problems wearing one word. Keeping them separate is the
whole design:

- **Ambient capture** — *mechanical, must-not-miss.* Every commit becomes a spine
  entry no matter how it was made. No agent needed, no context needed. Correctness =
  coverage.
- **Ambient expansion** — *agentic, context-rich.* Commit spine + the session's lived
  experience → narrative. Needs an agent that saw the work. Correctness = fidelity of
  the story while the context is warm.

The git post-commit hook (shipped) is a *good default for capture* and *nothing for
expansion*. The gaps we proved live: `core.hooksPath` overrides (now handled by
`doctor` + hooksPath-aware init), `--amend` double-fires, interactive rebase fires
`post-rewrite` not `post-commit`, and some GUI clients skip hooks. To go from "good"
to "can't-miss," capture needs a substrate below the hook.

## 1. Three candidate substrates

| Substrate | Catches | Misses | Infra | Right for |
|---|---|---|---|---|
| **git `post-commit` hook** (shipped) | normal + `--no-verify` commits | hooksPath\*, some GUIs, rebase, dedup on amend | none | default capture |
| **`.git/logs/HEAD` watcher daemon** | *everything* that moves HEAD — commit, amend, GUI, even reset/merge (reflog is appended regardless of hooks) | nothing, if running | a long-lived process (launchd/systemd) | can't-miss capture |
| **Claude Code / PAI hooks** (`Stop`, `SessionEnd`, `SubagentStop`) | work done *inside an agent session*, WITH the session context | any commit made outside an agent | CC/PAI already running | ambient **expansion** |

\* hooksPath now handled at install time; listed for completeness.

**Key realizations:**
- The reflog (`.git/logs/HEAD`) is the true event stream — git appends to it on every
  HEAD movement, hooks or not, GUI or CLI. A watcher tailing it and filtering for
  `commit`/`commit (amend)` lines is the tool-agnostic "can't-miss" capture. This is
  the real replacement for the old "#4 watcher."
- "Ambient Claude" isn't a first-class thing. The closest primitives are Claude Code
  **hooks** and **headless** (`claude -p`). So the agentic side lives in CC/PAI hooks;
  the mechanical side must live *below* any agent (the daemon), or it isn't ambient.

## 2. PAI / LifeOS: hreysi slots in, it doesn't bolt on

PAI (now **LifeOS**) runs *on top of Claude Code* and is built from the exact
primitives hreysi already uses. Its filesystem root is `~/.claude/`:

```
~/.claude/
├── Skills/<Name>/SKILL.md       # YAML frontmatter, "USE WHEN" triggers
├── hooks/{session-start,post-tool-use,stop,subagent-stop}/   # TS files
├── History/{Sessions,Learnings,Decisions,RawOutputs}/        # UOCS capture
└── agents/
```

Memory is "structured by purpose: WORK, KNOWLEDGE (typed graph), LEARNING,
RELATIONSHIP, OBSERVABILITY, STATE." The `Stop` hook already "extracts completion
summary, documents learnings."

The mapping is almost one-to-one:

| hreysi | PAI/LifeOS equivalent | Note |
|---|---|---|
| `expand` skill (`SKILL.md`, "USE WHEN") | a **Skill** in `~/.claude/Skills/` | *identical format* — drops in with zero adaptation |
| `buildlog/` commit spine | a **WORK / OBSERVABILITY** capture source | complements PAI's `History/` (git-commit stream vs session-transcript stream) |
| expansion trigger | a **`stop`/`subagent-stop` hook** | PAI's Stop hook literally "documents learnings" — hreysi's is the git-narrative specialization |
| (future) narrative → rules | **KNOWLEDGE** typed graph | where parked qortex interop lands |

So "plug into PAI" = ship hreysi's `expand` skill + a `stop` hook in PAI's layout, and
let hreysi's `buildlog/` be a capture source PAI's routing can read. hreysi stays
useful with *zero* PAI (plain git hook), and richer *with* it (agentic expansion).
**openclaw** is the natural home for the tool-agnostic watcher daemon (the capture
side that must live below any agent); PAI/CC owns the expansion side.

## 3. Recommended architecture (decoupled, matches directory-as-protocol)

```
                    CAPTURE (mechanical, tool-agnostic)          EXPANSION (agentic, context-rich)
 git commit ─┐
 GUI commit ─┼─▶ .git/logs/HEAD ─▶ hreysi watch ─┐              ┌─ CC/PAI stop|SessionEnd hook
 amend ──────┘        (daemon, can't-miss)        ├─▶ buildlog/ ─┤     runs `expand` vs un-expanded
 git commit ────────▶ post-commit hook ───────────┘  (protocol)  └─    commits, using live session
                          (default, 80%)                                 or persisted transcript
```

- **Capture layer** writes `## Commits`. Two implementations of the same contract:
  the git hook (default, zero infra) and the reflog watcher (`hreysi watch`, opt-in,
  can't-miss). Neither knows an agent exists.
- **Expansion layer** writes `## The Journey` / `## Improvements`. A CC/PAI hook,
  gated on "are there un-expanded commits?", runs the `expand` skill. Its lived
  experience is the live conversation (`Stop` can re-prompt) or the persisted
  transcript (`SessionEnd` reads the transcript file — robust, no re-prompt).
- They meet only at `buildlog/`. This is the decoupled solution, and it's the same
  boundary hreysi already ships.

## 4. Build phases

**Phase A — `hreysi watch` (can't-miss capture).**
- Tail `.git/logs/HEAD`; on new lines whose message starts `commit`/`commit (amend)`,
  run capture for that hash. Filter out `checkout`/`reset`/`rebase` noise; debounce
  rebases (many lines fast).
- Single-repo (`hreysi watch` in a repo) first; a global multi-repo daemon second.
- macOS `launchd` plist (and a systemd unit) so it survives logout. `hreysi watch`
  is the openclaw-hostable piece.
- Requires the **capture-dedup marker** below so the watcher and the hook don't
  double-log the same commit, and `--amend` doesn't duplicate.

**Phase B — the expansion hook (ambient expansion).**
- Ship a CC/PAI-compatible hook. Default trigger: **`SessionEnd`** (clean, one
  expansion per session, reads the transcript). Optional **gated `Stop`** for
  ultramarathon sessions (expand incrementally when N un-expanded commits accrue).
- Gate on a **last-expanded marker** (`buildlog/.hreysi/last_expanded` = HEAD at last
  expansion). Only fire when new commits exist since; update after.
- `hreysi init --ambient` installs *both* the git capture hook and the CC expansion
  hook; plain `hreysi init` stays capture-only.

**Phase C — PAI/LifeOS + openclaw packaging.**
- Publish the `expand` skill + `stop`/`SessionEnd` hook as a LifeOS-compatible bundle
  (its `SKILL.md` already conforms). Document dropping into `~/.claude/Skills/`.
- Home `hreysi watch` in openclaw as the tool-agnostic capture daemon.
- Optional, deferred: a local `hreysi` MCP server (`capture`, `doctor`, `expand-status`)
  so PAI's agentic routing can call it. **Defer** — a *hosted* MCP reopens the
  accounts/multi-tenancy question we parked; a local stdio MCP is fine but low value
  until the daemon + hook exist.

## 5. Markers & dedup (shared prerequisite)

Introduce `buildlog/.hreysi/`:
- `last_captured` — HEAD hash of the last captured commit. Capture skips if HEAD
  already logged (fixes the `--amend` double-fire and lets hook + watcher coexist).
- `last_expanded` — HEAD hash at last expansion. Gates the expansion hook.

This also cleanly resolves the queued **amend-dedup** item.

## 6. Decisions needed before building

1. **Watcher scope** — single-repo `hreysi watch` first (simpler, demoable), or jump
   to a global multi-repo daemon (more "LifeOS," more infra)? *Rec: single-repo first.*
2. **Expansion trigger** — `SessionEnd` only (clean), or `SessionEnd` + gated `Stop`
   (supports mid-session expansion for long sessions)? *Rec: SessionEnd default, add
   gated Stop behind a flag.*
3. **Lived-experience source for the hook** — persisted transcript (robust, works at
   SessionEnd) vs. live re-prompt (only via Stop). *Rec: transcript-based.*
4. **PAI coupling** — ship hreysi as a standalone tool that *happens to* conform to
   PAI's layout (keep it usable with zero PAI), vs. a PAI-first LifeOS skill bundle?
   *Rec: standalone-that-conforms; PAI is a distribution target, not a dependency.*

---

## Sources
- [Building a Personal AI Infrastructure (PAI) — December 2025](https://danielmiessler.com/blog/personal-ai-infrastructure-december-2025)
- [Building Your Own Personal AI Infrastructure](https://danielmiessler.com/blog/personal-ai-infrastructure)
- [danielmiessler/Personal_AI_Infrastructure (LifeOS) — GitHub](https://github.com/danielmiessler/Personal_AI_Infrastructure)
- [Announcing PAI 5.0 / LifeOS](https://danielmiessler.com/blog/announcing-pai-5-life-operating-system)
