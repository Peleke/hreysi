<div align="center">

# hreysi

### Ambient buildlog capture — every commit leaves a stone

[![Go](https://img.shields.io/badge/Go_1.26-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://go.dev/)
[![CI](https://img.shields.io/github/actions/workflow/status/Peleke/hreysi/ci.yml?branch=main&style=for-the-badge&logo=github&label=CI)](https://github.com/Peleke/hreysi/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/Peleke/hreysi?style=for-the-badge&logo=github&label=release)](https://github.com/Peleke/hreysi/releases)
[![Homebrew](https://img.shields.io/badge/Homebrew-Peleke%2Ftap-FBB040?style=for-the-badge&logo=homebrew&logoColor=white)](https://github.com/Peleke/homebrew-tap)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg?style=for-the-badge)](https://opensource.org/licenses/MIT)

**Commit like you always do. Your work journals itself.**

[Install](#install) · [Use](#use) · [What It Writes](#what-it-writes) · [The Idea](#the-idea) · [Build](#build-from-source)

</div>

---

## The Idea

Every commit is a decision you already made and already described. Then it evaporates — the message scrolls past, the context is gone, and next week you can't remember what the week even *was*.

**hreysi** catches it. One `init`, and from then on every `git commit` is appended to a dated markdown journal — no command to remember, no tool to reach for, no ceremony.

> *hreysi* (Old Norse) — a **cairn**: a pile of stones left to mark a path so the route stays legible to whoever comes later. Each commit drops a stone. The pile becomes the story.

It does exactly one thing, and it never gets in your way:

- **No wall.** hreysi installs a *non-blocking* `post-commit` hook. There is no pre-commit gate, no enforcement, no `BUILDLOG_COMMIT=1` bypass to memorize. Commit however you like — capture fires after, and a hiccup can never fail a commit.
- **Real time, not wall-clock.** Each block is stamped with git's committer date (`%cI`), so a capture fired from a hook still records the true commit time. Downstream gets an accurate event stream.
- **Zero dependencies.** A single static Go binary. No runtime, no venv, no Python version roulette.

## Install

```sh
curl -sSL https://raw.githubusercontent.com/Peleke/hreysi/main/install.sh | sh
```

Or with Homebrew:

```sh
brew install Peleke/tap/hreysi
```

Or grab a binary from [Releases](https://github.com/Peleke/hreysi/releases).

## Use

```sh
cd your-repo
hreysi init                 # scaffold buildlog/ + install the capture hook
git commit -m "feat: ..."   # → appended to buildlog/YYYY-MM-DD.md, automatically
```

That's the whole workflow. There is no step 3.

## What It Writes

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

One file per day. One block per commit. Timestamped, with the files it touched.

## The Directory *Is* the Product

`buildlog/` is the entire surface — and a deliberate decoupling boundary. hreysi **only ever writes** it; everything downstream **only ever reads** it:

```
git commit ──▶ hreysi ──▶ buildlog/*.md ──▶ ┌─ narrative expansion (the story of the week)
  (producer)                    (protocol)   ├─ content pipeline (posts, threads)
                                             └─ learning loop (mine the build narrative)
```

None of those consumers link back into hreysi. Build to the directory, not the tool — and any of them can be swapped, added, or automated later without touching capture.

## Ambient expansion (optional)

`hreysi init --ambient` also wires a Claude Code **SessionEnd** hook that, when
there are commits since the last expansion, runs the `expand` skill against the
session transcript — so the narrative gets written *while the context is warm*,
with nothing to remember. Gated so it never repeats work; best-effort and
non-fatal. Add `--ambient-stop` to also expand on long sessions.

## Using hreysi with LifeOS / PAI

hreysi is built from the same primitives as Daniel Miessler's
[LifeOS / PAI](https://github.com/danielmiessler/Personal_AI_Infrastructure) —
Skills, lifecycle hooks, and a filesystem capture layer — so it *slots in* rather
than integrating. `hreysi skills --global` drops the `expand` skill into
`~/.claude/Skills/`, and `hreysi init --ambient` wires capture + expansion per
repo. Full **agent-followable** setup (your Digital Assistant can run it):
[`docs/lifeos-integration.md`](docs/lifeos-integration.md).

## Commands

| Command | Does |
|---|---|
| `hreysi init` | Scaffold `buildlog/` and install the post-commit hook (honors `core.hooksPath`) |
| `hreysi capture` | Append HEAD to today's entry (run by the hook; also manual) |
| `hreysi watch` | Watch the reflog and capture every commit — any client, can't-miss |
| `hreysi doctor` | Verify capture is actually wired and will fire |
| `hreysi version` | Print version |
| `hreysi help` | Show help |

`hreysi init` installs into the directory git *actually* runs hooks from, so it
works even when `core.hooksPath` is overridden (husky, lefthook, …). Run
`hreysi doctor` any time to confirm capture is live — it's the one-command answer
to "is this thing on?"

### Can't-miss capture: `hreysi watch`

The git hook covers the vast majority of commits, but it can be bypassed
(`core.hooksPath`, some GUI clients, `--no-verify` edge cases). `hreysi watch`
tails the git **reflog** (`.git/logs/HEAD`), which git appends to on *every*
commit no matter how it's made — so capture becomes truly can't-miss. It's
idempotent with the hook (a commit is never logged twice) and handles `--amend`
by replacing the prior block. Run it in the foreground, or as a background
service (a sample `launchd`/`systemd` unit lives in `docs/`).

## Build from Source

```sh
git clone https://github.com/Peleke/hreysi
cd hreysi
go build -o hreysi .
go test ./...
```

Stdlib only — nothing to fetch.

---

<div align="center">

**Leave a legible trail. Let a later traveler read it.**

MIT © [Peleke Sengstacke](https://github.com/Peleke)

</div>
