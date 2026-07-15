---
name: article-runner
description: Fan out an APPROVED Campaign Brief's articles[] to the writer skills — the runner that turns a reviewed brief into article drafts. For each article brief, invokes article-draft (thought) or technical-tutorial (tutorial) and writes a draft into the portfolio; for tutorials it also scaffolds the sample-repo + TEACH.md and flags the package as pending. Idempotent: never re-drafts an article it already produced. Runs ONLY on status:approved — the human's gate before token-heavy drafting. Nothing publishes.
---

# hreysi article-runner

Take an **approved** Campaign Brief and fan its `articles[]` out to the writer skills,
producing draft articles in the portfolio. This is the step *after* the human review gate:
`digest-to-brief` writes the brief, a human approves it, and **this** spends the tokens to
draft. Consumes `articles[]` only — `beats[]` belong to linwheel's campaign mode, not here.

> This is the **expensive** step. Article drafts are the one thing we don't waste. So the
> runner is gated hard (approved-only) and idempotent (never re-drafts). Read both rules
> below before running it.

## Two hard gates (both protect article-draft spend)

1. **Approved-only.** If the brief's `status` is not `approved`, STOP. Do nothing. Tell the
   user the brief is still a draft and they must review + approve it first. The runner never
   flips `status` itself — approval is the human's decision and their token authorization.
2. **Idempotent.** The brief records which article ids it has already drafted in a
   `drafted:` frontmatter list. The runner drafts **only** article briefs whose id is NOT in
   that list, and appends each id as it completes. Re-running the runner on the same brief
   drafts nothing already done. A crash mid-run re-runs only the unfinished articles.

## Prerequisites

- The writer skills must be installed: `article-draft` and `technical-tutorial` (from
  `brunnr/skills`). If either is missing, stop and say which — the runner does not draft
  prose itself, it orchestrates the writers.
- A portfolio target for drafts (default `portfolio/articles/`, per the article handoff).

## Inputs

- **An approved Campaign Brief** — `<vault>/Campaigns/YYYY-Www.md` with `status: approved`.
  Its `articles[]` is the work-list; each entry carries everything the matching writer's
  Step 0 needs (built by `digest-to-brief` from the corpus).

## Steps

1. **Gate.** Read the brief. If `status != approved` → stop (gate 1). Read the `drafted:`
   list (empty if absent).
2. **Select.** The work-list is every `articles[]` entry whose `id` is not in `drafted:`.
   If empty → report "all articles already drafted" and stop.
3. **Pick the next number.** Scan the portfolio for the highest `NN-` prefix; new drafts
   continue the sequence (`NN-<slug>-DRAFT.md`), matching `portfolio/articles/README.md`.
4. **Draft each selected article, in `order` (article beats before their orbiting posts):**

   **`kind: thought` → `article-draft`.** Hand it the article brief's `editorial_map`
   (claims + evidence — its Step 0 "editorial map"), `war_stories`, and `voice`. It produces
   the prose. Write `portfolio/articles/NN-<slug>-DRAFT.md`.

   **`kind: tutorial` → `technical-tutorial` + package scaffold.**
   - Invoke `technical-tutorial` with `capability`, `prerequisites`, `companion_artifact`,
     `war_stories`, `voice` → write `portfolio/articles/NN-<slug>-tutorial-DRAFT.md`.
   - **Scaffold the package** (do NOT build it fully — that is the later
     tutorial-package-builder):
     - `mkdir sample-repos/<slug>/` (per `sample_repo`'s extraction intent).
     - Write `sample-repos/<slug>/TEACH.md` as a **stub** — the oscillation skeleton from
       `docs/tutorial-package-spec.md`: Starting State, the DO→NOTICE→CODE→NAME turns, and
       the instructor→student HANDOFF points left as TODOs.
     - Write `sample-repos/<slug>/PACKAGE-PENDING.md` flagging what's left: extract the
       clean-room runnable repo, stage `learn/step-N`, wire the mistake-ledger port.
   - The prose ships now; the package is explicitly pending, not silently missing.

5. **Record.** After each article completes, append its `id` to the brief's `drafted:`
   frontmatter list (idempotency). Do this per-article, not at the end, so a crash is
   recoverable.
6. **Report.** List what was drafted (paths), what was scaffolded-and-pending (tutorials),
   and what was skipped (already drafted). Remind the user: these are **drafts** in the
   portfolio; nothing published.

## Contract

- **Never runs on a draft brief.** Approval is the gate and the token authorization.
- **Never re-drafts.** The `drafted:` list is the guard; respect it absolutely. Re-spending
  on an already-drafted article is the exact waste this skill exists to prevent.
- **Never publishes.** Output is draft files in the portfolio. Publishing an article is a
  separate human step (a PR in the portfolio repo, or wherever the human ships).
- **Doesn't touch `beats[]`.** Those are linwheel's. The runner is the long-form pipeline.
- **Doesn't fully build tutorial packages.** It drafts prose and scaffolds; the clean-room
  repo extraction, `learn/step-N` staging, and mistake-ledger wiring are the
  tutorial-package-builder's job (later). Flag them, don't fake them.

## Gotchas

- **The `drafted:` list is per-article, not per-brief.** A brief with 3 articles, 1 drafted,
  re-runs the other 2 only. Don't treat "brief touched" as "brief done."
- **A tutorial's prose can ship before its package exists.** That's intended — the article
  is useful standalone; the interactive package is an enhancement. Just never claim the
  package is done when only the stub exists.
- **If a writer skill fails on one article, keep going** on the others and report the
  failure — don't let one bad article block a whole approved brief. The failed one stays out
  of `drafted:`, so a re-run retries just it.
- **Match the writer to the kind, not the other way.** A `kind: tutorial` with no
  `capability` field is a malformed brief — stop and report it, don't guess.

## Examples

- "run the approved campaign brief" → read `Campaigns/2026-W29.md`; if approved, draft the
  undrafted articles (thought via article-draft, tutorial via technical-tutorial + scaffold),
  record each in `drafted:`, report the portfolio paths.
- "run it again" → everything's in `drafted:`, so it drafts nothing and says so. No tokens
  spent re-drafting.
- "the brief is still a draft" → runner refuses, tells the user to review and approve first.
