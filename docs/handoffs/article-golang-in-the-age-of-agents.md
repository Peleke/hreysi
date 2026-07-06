# Handoff: "Learning Golang in the Age of Agents" article

> **For the writing agent.** A brief, not a draft. The worked example is the real `hreysi`
> codebase, built in one session on 2026-07-06 (see **Sources**). Use the actual code as the
> teaching material — quote it. Don't fabricate code that isn't in the repo.

## One-line thesis

You can learn a language *by directing and reading an agent's implementation of a real tool*,
and Go is unusually well-suited to this: it's small, explicit, and readable, so the code an
agent produces is code you can actually learn *from*. This article teaches Go's core moves
through hreysi — a ~300-line, stdlib-only CLI — and, underneath, makes a claim about how
language-learning changes when an agent writes the first draft.

## Audience & promise

- **Audience:** engineers fluent in another language (Python/TS/etc.) who keep meaning to
  "learn Go," and people curious what agent-assisted learning actually looks like in practice.
- **Promise:** after reading, they understand the Go concepts that carry a real CLI, and they
  have a *method* — read the agent's code, understand each idiom, then you can direct the next
  change with intent instead of vibes.

## Structure

1. **Frame: the tool, and why Go fit it.** hreysi captures every git commit into a dated
   journal. The workload is I/O-bound orchestration (shell out to git, append a file), so the
   real reason to pick Go wasn't speed — it was a **single static, zero-dependency binary**.
   Make the meta-point early: pick the language for the *artifact*, and Go's artifact is a
   great one to learn on because there's no runtime/venv ceremony between you and the code.
2. **The Go tour, driven by hreysi's actual code.** Each concept anchored to a real file:
   - **Packages & internal/** — how the repo is split (`internal/gitx`, `internal/entry`,
     `internal/scaffold`, `internal/skillpack`) and what `internal/` means (import-scoped to
     the module). Small packages with one job each = readable, testable.
   - **Errors as values** — `if err != nil { return … }`, wrapping with `%w`, no exceptions.
     Show the git wrapper and how failures propagate up to `main` and become exit codes.
   - **Shelling out** — `os/exec` (`exec.Command`, `cmd.Dir`, `.Output()`) in `gitx.go`. Why
     hreysi shells to the git binary instead of linking a git library (dependency-free, git
     is always present in the repo it's hooking).
   - **Strings & building output** — `strings.Builder`, `strings.Index`/`TrimRight`/`TrimLeft`
     in `entry.go`, including the `insertCommit` function (inserting a commit block *inside*
     the `## Commits` section, before narrative sections). Good, concrete string-surgery.
   - **The filesystem** — `os.ReadFile`/`WriteFile`/`MkdirAll`, `path/filepath.Join` for
     portable paths, file modes (`0o755` for the executable hook).
   - **`//go:embed`** — embedding the `skills/` directory into the binary and walking it with
     `io/fs` (`fs.WalkDir`, `fs.ReadFile`) in `skillpack.go`. This is a standout Go feature:
     ship assets *inside* the binary, drop them on `init`. Great "oh, that's clean" moment.
   - **`main` as a dispatcher** — the `switch os.Args[1]` command router, `-ldflags
     "-X main.version=…"` for injecting the version at build time. Show the release wiring.
   - **Testing that proves something** — hreysi's tests create **real temp git repos**
     (`t.TempDir()`, real `git init`/`commit`) and assert on real captured output. Use this to
     teach `testing` basics *and* a values point: mocked-out tests of a git tool prove nothing;
     the only test that proves capture works is one that actually commits. (This is a genuine
     conviction in the project — carry it.)
3. **The agent angle (the reason for the title).** Weave, don't bolt on:
   - What the human actually held: the *shape* (packages, the decoupling boundary, "insert
     inside the section, not at EOF"), the *tradeoffs* (Go vs Rust, why), and the *taste*
     (name the real reason, keep dead ends). What the agent held: the idiom-level syntax.
   - The learning loop that emerges: you read the agent's Go, understand each idiom against a
     concept you already know from another language, and now you can direct the next edit
     precisely. The bug we caught live is the perfect example: expansion added sections after
     `## Commits`, so the naive "append at EOF" put new commits *after* the narrative — a
     human reading the code spotted the structural error, and the fix (`insertCommit`) plus a
     real-git regression test followed. That's the loop: read → understand → direct → verify.
   - Honest boundary: agent-assisted learning gets you *reading fluency* and *directing
     ability* fast; *writing fluency* (typing idiomatic Go from scratch) still takes reps.
     Say so.
4. **Close.** Go is a good first language for the agent era precisely because its code is
   legible — the agent's output is a teacher, not a black box. Learning-by-directing is real,
   and it starts with being able to *read* what you asked for.

## Concrete anchors (real, from the repo)

- ~300 lines, **stdlib only** (no third-party deps — `go.mod` has none). Single binary.
- Files to quote: `main.go` (dispatch, embed, ldflags), `internal/gitx/gitx.go` (exec, errors),
  `internal/entry/entry.go` (`insertCommit`, string surgery, fs), `internal/skillpack/skillpack.go`
  (`//go:embed`, `io/fs`), the `_test.go` files (real-git fixtures).
- The live bug + fix: `insertCommit` and `TestAppendLandsInsideCommitsAfterExpansion`.

## Tone & constraints

- House voice: punchy, honest, em-dashes; teach by showing real code and real output.
- Don't teach Go exhaustively (no goroutines/channels — hreysi has none; don't invent a
  concurrency section). Teach exactly the Go that *this tool* uses. That constraint is a
  feature: it keeps the article grounded and finishable.
- The agent-learning thread is the differentiator — keep it present but let the code carry it.

## Sources (read these first)

- **Primary:** `hreysi/buildlog/2026-07-06.md` — `## The Journey` (Go-vs-Rust decision, the
  live bug) and `## Improvements → Tooling` (compiled-for-distribution, `//go:embed`).
- The hreysi source files listed above.
- `hreysi/README.md` for voice + the tool's premise.
