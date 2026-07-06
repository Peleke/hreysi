# hreysi

**Ambient buildlog capture.** Every git commit, appended to a dated journal.
No ceremony, nothing to remember.

> *hreysi* (Old Norse) — a cairn; a pile of stones left to mark a path so the
> route stays legible to whoever comes later. Each commit drops a stone.

---

## Install

```sh
curl -sSL https://raw.githubusercontent.com/Peleke/hreysi/main/install.sh | sh
```

Or grab a binary from [Releases](https://github.com/Peleke/hreysi/releases).

## Use

```sh
cd your-repo
hreysi init                 # scaffold buildlog/ + install the capture hook
git commit -m "feat: ..."   # → appended to buildlog/YYYY-MM-DD.md, automatically
```

That's it. `hreysi init` installs a **non-blocking** `post-commit` hook. There
is no pre-commit wall, no enforcement, no bypass to remember — capture is a
side-effect of committing, and a hiccup can never fail a commit.

## What it writes

```markdown
# 2026-07-06

## Commits

### `a1b2c3d` — feat: add capture command
_2026-07-06T21:51:54-04:00_

Files:
- `main.go`
- `internal/entry/entry.go`
```

Dated file per day, one block per commit, each stamped with the **real commit
time** (git's committer date, not wall-clock — so a hook-fired capture is still
accurate).

## The directory is the product

`buildlog/` is the entire surface and the decoupling boundary. hreysi only ever
**writes** it; everything downstream only ever **reads** it:

- a narrative-expansion skill that turns commit stubs into the story of the week
- a content pipeline that reshapes that story into posts
- a learning loop that mines the build narrative

None of them link back into hreysi. Build to the directory, not to the tool.

## Commands

| Command | Does |
|---|---|
| `hreysi init` | Scaffold `buildlog/` and install the post-commit hook |
| `hreysi capture` | Append HEAD to today's entry (run by the hook; also manual) |
| `hreysi version` | Print version |
| `hreysi help` | Show help |

## Build from source

```sh
go build -o hreysi .
go test ./...
```

MIT licensed.
