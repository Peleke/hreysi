# Using hreysi inside LifeOS / PAI

> **This page is written to be executed by an agent.** If you're a LifeOS/PAI
> Digital Assistant setting hreysi up for your human, follow the steps verbatim.
> hreysi conforms to LifeOS's primitives (Skills + hooks + a filesystem capture
> layer), so integration is a few commands — no hand-editing.

## What hreysi adds to LifeOS

| hreysi piece | LifeOS home | Role |
|---|---|---|
| `expand` skill (`SKILL.md`) | `~/.claude/skills/expand/` | narrate git commits into the story of the work |
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
hreysi skills --global                 # writes ~/.claude/skills/expand/SKILL.md
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

**4. (Optional) content generation — opt in.** If the human wants their week turned
into LinkedIn drafts, add the `reshape` skill (needs the linwheel MCP configured):
```sh
hreysi skills --global --linwheel      # adds the reshape weekly-digest skill
# or per repo:  hreysi init --ambient --linwheel
```
Leave `--linwheel` off and hreysi never touches content — capture + narrative only.

**5. (Optional) can't-miss capture.** If the human uses GUI git clients or repos
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

## The daily rhythm (what the human actually experiences)

Once installed, the loop is quiet by design — most of it is invisible:

| When | What happens | Automatic? |
|---|---|---|
| Every commit | Captured to today's `buildlog/` entry (the spine) | ✅ silent |
| End of each session | The day's commits get narrated into `## The Journey` + `## Improvements` from the transcript | ✅ silent (SessionEnd hook) |
| **Weekly, on demand** | Human (or a scheduled routine) runs the `reshape` skill → the week's narrative is scanned, triaged, and turned into LinkedIn **drafts** in the linwheel dashboard | ⚙️ **manual / opt-in** |
| Whenever | Human reviews / approves / schedules those drafts **in linwheel** | 👤 human |

The deliberate part: **content never auto-generates and never auto-publishes.**
Capture and narrative are ambient; turning that into posts is a weekly choice
(fire the `reshape` skill, or wire a weekly `schedule`/cron to run it). Drafts land
in linwheel and stop there — approval and scheduling stay in linwheel's editor.

So the honest one-liner for the human: *"You commit as usual. Your work journals
and narrates itself. Once a week, if you want, it becomes a stack of LinkedIn drafts
waiting for your yes/no."*

## Uninstall / opt-out
- Remove `~/.claude/skills/expand/` and the SessionEnd entry in `.claude/settings.json`.
- Delete `.git/hooks/post-commit` (or the entry hreysi appended) to stop capture.
