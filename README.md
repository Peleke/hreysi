<div align="center">

# hreysi

### Ambient buildlog capture — every commit leaves a stone

[![Go](https://img.shields.io/badge/Go_1.26-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://go.dev/)
[![CI](https://img.shields.io/github/actions/workflow/status/Peleke/hreysi/ci.yml?branch=main&style=for-the-badge&logo=github&label=CI)](https://github.com/Peleke/hreysi/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/Peleke/hreysi?style=for-the-badge&logo=github&label=release)](https://github.com/Peleke/hreysi/releases)
[![Homebrew](https://img.shields.io/badge/Homebrew-Peleke%2Ftap-FBB040?style=for-the-badge&logo=homebrew&logoColor=white)](https://github.com/Peleke/homebrew-tap)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg?style=for-the-badge)](https://opensource.org/licenses/MIT)

**Commit like you always do. Your work journals itself — then becomes the story of your week.**

[Install](#install) · [Quick start](#quick-start) · [How it works](#how-it-works) · [Capture](#capture-mechanical) · [Expand](#expand-agentic) · [Feed it anywhere](#feed-it-anywhere) · [LifeOS](#using-hreysi-with-lifeos--pai) · [Commands](#commands)

</div>

---

## The idea

Every commit is a decision you already made and already described — then it evaporates. The message scrolls past, the reasoning is gone, and next week you can't remember what the week even *was*.

**hreysi** catches it. One `init`, and from then on every `git commit` is appended to a dated markdown journal — no command to remember, no tool to reach for, no ceremony. Then, when you want, an agent turns that raw commit log into the actual **story** of the work: what you tried, what broke, what you learned.

> *hreysi* (Old Norse) — a **cairn**: a pile of stones left to mark a path so the route stays legible to whoever comes later. Each commit drops a stone. The pile becomes the story.

Two layers, kept deliberately separate:

- **Capture is mechanical.** A commit → a dated entry. No agent, no thought, can't be forgotten. This is the *spine*.
- **Expansion is agentic.** The spine + the lived experience of the session → narrative. Commit metadata is an index; the story lives in the work and decays fast, so you graft it on while the context is warm. This is the *body*.

They meet in one place — the `buildlog/` directory — and nothing downstream needs to know hreysi exists. That directory is the whole product and the reason a content pipeline, a portfolio generator, or a learning loop can all read your work without touching the tool.

## Install

```sh
brew install Peleke/tap/hreysi
```

or, with no Homebrew and no runtime to install:

```sh
curl -sSL https://raw.githubusercontent.com/Peleke/hreysi/main/install.sh | sh
```

or grab a single static binary from [Releases](https://github.com/Peleke/hreysi/releases). No Python, no venv, no dependencies — one file.

## Quick start

```sh
cd your-repo
hreysi init                 # scaffold buildlog/ + install the capture hook
git commit -m "feat: ..."   # → appended to buildlog/YYYY-MM-DD.md, automatically
hreysi doctor               # confirm capture is actually live
```

That's the whole workflow. There is no step 4.

### What it writes

```markdown
# 2026-07-06

## Commits

### `4689e74` — feat: initial app
_2026-07-06T18:49:41-04:00_

Files:
- `README.md`
- `app.py`

### `1a297b3` — fix: add config constant
_2026-07-06T18:49:42-04:00_

Files:
- `app.py`
```

One file per day, one block per commit, each stamped with the **real commit time** (git's committer date, not wall-clock — so a hook-fired capture is still accurate). After you [expand](#expand-agentic), the same file grows a `## The Journey` and `## Improvements` section beneath the commits.

## How it works

```
                CAPTURE (mechanical)                     EXPAND (agentic)
 git commit ─┐
 GUI commit ─┼─▶ buildlog/YYYY-MM-DD.md ──────────────▶  an agent reads the spine
 amend ──────┘   ## Commits  (the spine)                 + this session's context,
                     ▲                                    writes ## The Journey +
              hook · watcher                              ## Improvements (the body)
                                        └──────────────▶  feed the file anywhere:
                                          buildlog/         LinkedIn, portfolio,
                                        (the protocol)      a learning loop, …
```

## Capture (mechanical)

Capture is a side-effect of committing. There is **no pre-commit wall**, no enforcement, no bypass to memorize — `hreysi init` installs a *non-blocking* `post-commit` hook, so a hiccup can never fail a commit, and you commit however you already do.

**It installs where git actually looks.** `hreysi init` writes into the directory git *really* runs hooks from (`git rev-parse --git-path hooks`), so capture works even when `core.hooksPath` is overridden by husky, lefthook, or the pre-commit framework — the #1 way naive hooks silently stop firing.

**`hreysi doctor` — is this thing on?** One command verifies capture is wired and will fire (hook present, points at hreysi, executable, in the effective hooks dir). Exit 0 healthy, exit 1 if capture won't fire, with the fix.

```
$ hreysi doctor
hreysi doctor — /path/to/repo
  hooks dir: /path/to/repo/.git/hooks
  ✓ post-commit hook present
  ✓ hreysi capture wired
  ✓ hook executable
  ✓ journal directory — buildlog/

capture is live — every commit will be journaled.
```

### `hreysi watch` — can't-miss capture

The git hook covers the vast majority of commits, but it *can* be bypassed — some GUI clients skip hooks, `core.hooksPath` can point elsewhere, edge cases exist. When you want a guarantee, `hreysi watch` tails the git **reflog** (`.git/logs/HEAD`), which git appends to on *every* commit no matter how it's made. Capture becomes truly can't-miss.

```sh
hreysi watch     # foreground; captures every commit from any client
```

It's idempotent with the hook (a commit is never logged twice) and handles `--amend` by replacing the prior block, not duplicating it. Run it as a background service — a sample `launchd`/`systemd` unit is in [`docs/watch-service.md`](docs/watch-service.md).

## Expand (agentic)

Capture gives you the spine. **Expansion gives you the story** — and the story is what makes the journal worth reading and worth reshaping into content. `hreysi init` drops an `expand` skill into `.claude/skills/`.

**Manual.** In Claude Code, in the repo, run the skill (`/expand`, or "expand today's buildlog"). It reads the `## Commits` spine plus *this session's* lived experience — what you attempted, what fought back, the decisions and dead ends — and writes `## The Journey` + `## Improvements`, never touching `## Commits`. Great for demos and for firing mid-session on a long haul.

**Ambient.** `hreysi init --ambient` also wires a Claude Code **SessionEnd** hook that runs `expand` automatically when there are commits since the last expansion — so the narrative is written while the context is warm, with nothing to remember. It's gated (never repeats work), transcript-based, and non-fatal. Add `--ambient-stop` to also expand on long ("ultramarathon") sessions.

Why the split? The lived experience — the *why*, the friction, the aha — isn't in git. It's in the session, and it evaporates when the session ends. So capture must be mechanical and always-on; narrative must be grafted on by an agent, close to the work.

## Feed it anywhere

`buildlog/` is the entire surface and a deliberate decoupling boundary: hreysi **only ever writes** it, and everything downstream **only ever reads** it. Nothing links back into the tool.

The headline consumer: **turn your week into LinkedIn.** An expanded entry — the real story, dead ends and all — is exactly the source text a tool like [linwheel](https://www.linwheel.io) reshapes into posts. You already generated the content; you just committed it.

This is **opt-in and manual**. `hreysi init --linwheel` (or `hreysi skills --linwheel`) adds a `reshape` skill: a **weekly digest** you fire on demand that scans the week's narrative, triages the post-worthy threads, and drops LinkedIn **drafts** into your linwheel dashboard for review. Nothing auto-generates, nothing auto-publishes — leave `--linwheel` off and hreysi never touches content. The same directory feeds a portfolio generator, a changelog, or a learning loop just as easily. Build to the directory, not the tool.

## Using hreysi with LifeOS / PAI

hreysi is built from the same primitives as Daniel Miessler's [LifeOS / PAI](https://github.com/danielmiessler/Personal_AI_Infrastructure) — Skills, lifecycle hooks, and a filesystem capture layer — so it *slots in* rather than integrating.

```sh
hreysi skills --global      # drop the expand skill into ~/.claude/skills/ (PAI layout)
hreysi init --ambient       # per repo: capture + expansion, wired
```

Your commit stream (`buildlog/`) becomes a WORK/OBSERVABILITY source that complements PAI's session-transcript `History/`; the `expand` skill and SessionEnd hook are exactly the primitives PAI already runs. Full **agent-followable** setup — your Digital Assistant can execute it — in [`docs/lifeos-integration.md`](docs/lifeos-integration.md). hreysi also runs perfectly with **zero** LifeOS; it just gets richer inside one.

## Commands

| Command | Does |
|---|---|
| `hreysi init` | Scaffold `buildlog/`, install the capture hook (honors `core.hooksPath`), drop the `expand` skill. `--ambient` wires a SessionEnd expansion hook; `--ambient-stop` adds Stop; `--no-skill` skips the skill |
| `hreysi capture` | Append HEAD to today's entry — run by the hook; also manual/backfill. Idempotent; replaces on `--amend` |
| `hreysi watch` | Tail the reflog and capture every commit from any client — can't-miss |
| `hreysi doctor` | Verify capture is wired and will fire |
| `hreysi skills` | Install the bundled skills (`--global` → `~/.claude/skills/` for LifeOS/PAI) |
| `hreysi version` | Print version |
| `hreysi help` | Show help |

## Build from source

hreysi is ~300 lines of Go, **stdlib only** — nothing to fetch, and it compiles to a single static binary.

```sh
git clone https://github.com/Peleke/hreysi
cd hreysi
go build -o hreysi .
go test ./...          # real git repos, no mocks
```

Releases are cut by tag: GitHub Actions runs the tests, [goreleaser](https://goreleaser.com) builds every platform, and the binaries + Homebrew formula + `install.sh` all publish from that one pipeline. Every install you can do comes from an artifact the pipeline produced — never a local build.

---

<div align="center">

**Leave a legible trail. Let a later traveler read it.**

MIT © [Peleke Sengstacke](https://github.com/Peleke)

</div>
