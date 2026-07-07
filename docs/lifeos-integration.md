# Using hreysi inside LifeOS / PAI

> **This page is written to be executed by an agent.** If you're a LifeOS/PAI
> Digital Assistant setting hreysi up for your human, follow the steps verbatim.
> hreysi conforms to LifeOS's primitives (Skills + hooks + a filesystem capture
> layer), so integration is a few commands — no hand-editing.

## What hreysi adds to LifeOS

| hreysi piece | LifeOS home | Role |
|---|---|---|
| `expand` skill (`SKILL.md`) | `~/.claude/Skills/expand/` | narrate git commits into the story of the work |
| commit capture (`buildlog/`) | a **WORK / OBSERVABILITY** source per repo | complements `History/` (git-commit stream vs session-transcript stream) |
| expansion hook | a **SessionEnd** hook | LifeOS's Stop hook already "documents learnings"; this specializes it to git narrative |

hreysi runs fine with **zero** LifeOS (plain git hook). These steps make it *native*
to a LifeOS install.

## Steps (run these)

**1. Install the binary.**
```sh
brew install Peleke/tap/hreysi        # or: curl -sSL https://raw.githubusercontent.com/Peleke/hreysi/main/install.sh | sh
hreysi version
```

**2. Register the `expand` skill globally in LifeOS.**
```sh
hreysi skills --global                 # writes ~/.claude/Skills/expand/SKILL.md
```
It uses the same `SKILL.md` + "use when" routing LifeOS expects, so LifeOS's router
picks it up at startup.

**3. Turn on capture for each repo the human works in.** For every repo:
```sh
cd /path/to/repo
hreysi init --ambient                  # capture hook + SessionEnd expansion hook + project skill
hreysi doctor                          # confirm capture is live
```
`--ambient` wires a **SessionEnd** hook that, when there are un-expanded commits,
runs the `expand` skill against the session transcript. Add `--ambient-stop` to also
expand on `Stop` for long ("ultramarathon") sessions.

**4. (Optional) can't-miss capture.** If the human uses GUI git clients or repos
with `core.hooksPath` overrides, run the reflog watcher as a background service so no
commit is ever missed:
```sh
hreysi watch                           # foreground; see docs/watch-service.md for launchd/systemd
```

## How it behaves inside LifeOS

- **Capture** is mechanical and always-on (git hook, or `hreysi watch`). It writes
  only the `## Commits` spine. No agent needed.
- **Expansion** is agentic. At SessionEnd, if commits are un-expanded, the hook asks
  an agent to run `expand`, which writes `## The Journey` + `## Improvements` from the
  transcript and stamps `buildlog/.hreysi/last_expanded`. Gated so it never repeats
  work.
- The `buildlog/` directory is the seam. LifeOS skills, a content pipeline (e.g.
  linwheel), or a KNOWLEDGE-graph importer can all read it without touching hreysi.

## Uninstall / opt-out
- Remove `~/.claude/Skills/expand/` and the SessionEnd entry in `.claude/settings.json`.
- Delete `.git/hooks/post-commit` (or the entry hreysi appended) to stop capture.
