# Handoff: "Homebrew How-To" article

> **For the writing agent.** This is a brief, not a draft. Your job is to turn it into
> the article. Everything factual below actually happened on 2026-07-06 while building
> `hreysi`; the primary source is that day's buildlog entry (see **Sources**). Do not
> invent events — if you need a detail that isn't here or in the sources, flag it.

## One-line thesis

You do not need to understand Homebrew's internals to ship a tool through it. You need
to know the **language of the concepts** — tap, formula, cask, bottle, the release
pipeline that feeds them — well enough to direct an agent to wire it up. This article
teaches that vocabulary by walking a real distribution loop we built end-to-end in one
session.

## Audience & promise

- **Audience:** developers who can build a CLI but have never *distributed* one; people
  who bounce off Homebrew because every guide assumes you already know what a "tap" is.
- **Promise:** by the end they can (a) reason about *which* distribution channel fits a
  given artifact, and (b) hand an agent a correct, specific instruction to set up a
  Homebrew tap — without reading the Homebrew source.

## The spine of the piece (structure)

1. **The rebuild story as the frame.** Open on the concrete problem: we had a single Go
   binary (`hreysi`) and wanted "install → it just works" for a live demo. Use the real
   arc — a static binary, then the question "how does a stranger get this?"
2. **Distribution tradeoffs — the decision, named honestly.** This is the intellectual
   core. Lay out the channels and when each fits:
   - **Language package managers** (pipx/pip, npm, cargo): great when the user already
     has that runtime; a *dependency* otherwise. We explicitly left Python/pipx because
     it drags Python 3.14 + a venv behind it.
   - **`curl -sSL … | sh`** (install script off GitHub Releases): zero prior tooling,
     works on any machine, "a bit sexier" for a technical audience. Needs only the repo's
     built-in `GITHUB_TOKEN`. This was our *primary* path and the most bulletproof.
   - **Homebrew**: the nicest UX on macOS/Linux (`brew install org/tap/tool`, upgrades,
     uninstall) — but it needs *more* moving parts (a second repo, a cross-repo token).
   - The lesson: **compile for distribution (single static binary), not for speed.** Our
     workload was I/O-bound git-shelling; "blazing speed" was a red herring. Name the real
     reason you reach for Go/Rust — it's the artifact shape, not the benchmark.
3. **How Homebrew is actually structured (the vocabulary).** Teach just enough:
   - **Tap** = a git repo named `homebrew-<name>` that Homebrew knows how to add
     (`brew tap org/name` → `github.com/org/homebrew-name`). It's just a folder of formulae.
   - **Formula** = a Ruby file (`Formula/hreysi.rb`) describing where to download the
     binary per OS/arch, the sha256, and how to install it. Show the real generated
     formula from the piece (the `on_macos`/`Hardware::CPU.arm?` branches, `bin.install`).
   - **Formula vs cask** = binary/source formula vs. pre-built app bundle; for a CLI you
     want a formula. (One sentence; don't rabbit-hole.)
   - **Bottle** = a pre-compiled formula; note it exists, say we didn't need it (we ship a
     pre-built binary via the formula's `url`), and move on.
   - **goreleaser** as the thing that generates all of the above from one `.goreleaser.yaml`
     on a version tag — the reader never hand-writes the Ruby.
4. **The one non-obvious gotcha, told as a gotcha.** The built-in Actions `GITHUB_TOKEN`
   can push to the *current* repo but **not to a different repo** — so publishing a formula
   to a separate tap repo needs its own token secret. We proved this: releases + install.sh
   worked with `GITHUB_TOKEN` alone; the tap required creating `Peleke/homebrew-tap` and a
   `HOMEBREW_TAP_GITHUB_TOKEN` secret. Include the security footnote: we used the broad `gh`
   OAuth token to move fast, and the *correct* posture is a fine-grained PAT scoped to just
   the tap. This "move fast, then name the debt" beat is on-brand and honest.
5. **The payoff, shown.** The proof sequence: tag `v0.1.1` → GitHub Actions + goreleaser →
   formula lands in the tap → `brew install Peleke/tap/hreysi` → binary in `/opt/homebrew/bin`.
   Emphasize the principle we insisted on: **prove the loop by installing from the
   published artifact, not a local build.** A green CI badge is not proof; `brew install`
   working is.
6. **The turn — "how to do this with an agent."** The summary/close. The reader didn't read
   Homebrew's docs; they learned the *concepts* (tap/formula/token-scope/pipeline) and that
   was enough to direct the work. Make the explicit claim: in the age of agents, the
   leverage is in holding the right conceptual vocabulary, not memorizing the Ruby DSL.
7. **Seed the follow-up.** Close by naming the deeper article this implies: **the difference
   between the knowledge needed to *build* something and the knowledge needed to *use* it.**
   Homebrew is a perfect specimen — you can *use* it fluently (tap/formula/brew install)
   with near-zero knowledge of how bottles are compiled or how the client resolves
   dependencies. One or two sentences; don't write that article here, just plant it.

## Concrete facts to anchor on (all real, from today)

- Tool: `hreysi`, a Go CLI. Repo `github.com/Peleke/hreysi` (public, MIT).
- Pipeline: `.github/workflows/release.yml` runs `goreleaser release --clean` on `v*` tags;
  `.goreleaser.yaml` defines cross-platform builds (linux/darwin × amd64/arm64), a `brews:`
  block targeting `Peleke/homebrew-tap`, and archive naming `hreysi_<version>_<os>_<arch>`.
- `install.sh` resolves the latest release tag via the GitHub API and downloads the matching
  tarball — the `curl | sh` path.
- Tap repo `Peleke/homebrew-tap` created fresh; secret `HOMEBREW_TAP_GITHUB_TOKEN` on the
  hreysi repo carries the push token.
- Versions shipped: v0.1.0 (releases + install.sh), v0.1.1 (added Homebrew), v0.1.2.
- Proven live: `brew install Peleke/tap/hreysi` → `/opt/homebrew/bin/hreysi` → `hreysi 0.1.2`.

## Tone & constraints

- House voice (see hreysi/README.md, cadence-planner, interlinear): punchy, em-dash prose,
  honest about tradeoffs and dead ends. Empirical — show the commands and their output.
- Don't over-teach Homebrew internals; the whole point is you *don't* need them. Stay at the
  vocabulary + pipeline level.
- Keep the security caveat (broad token → fine-grained PAT) — it's part of the honesty.

## Sources (read these first)

- **Primary:** `hreysi/buildlog/2026-07-06.md` — `## The Journey` (the distribution loop, the
  token gap, the proof) and `## Improvements → Tooling/Process/Gotchas`.
- `hreysi/.goreleaser.yaml`, `hreysi/.github/workflows/release.yml`, `hreysi/install.sh`.
- The generated formula: `github.com/Peleke/homebrew-tap` → `Formula/hreysi.rb`.
