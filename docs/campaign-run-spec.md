# `hreysi campaign` — firing the runner, mechanically

> **Status:** design + plan. Not built. Pick-up-later spec for turning the manual
> `article-runner` skill into a Go-orchestrated subcommand that manages the article-draft
> fan-out. Downstream of `campaign-brief-spec.md`, `digest-to-brief`, and the
> `article-runner` skill; read those first.

## The problem

Today you fire the runner **manually**: open a Claude Code session, invoke the
`article-runner` skill, and it drafts the approved brief's articles. That works, but it has
three costs:

1. **You are the orchestrator.** You carry an approved brief into a session and babysit N
   drafts. That is the human-as-integration-layer pattern — the thing the pipeline exists to
   remove, applied to the pipeline itself.
2. **N articles draft in one context.** A single agent session drafting three articles
   shares one context window; later drafts inherit the drift of earlier ones. Isolation is
   better — one fresh agent per article.
3. **It can't be scheduled or backgrounded.** A skill invocation needs a live session. There
   is no `cron`-able way to say "when I approve a brief, draft it."

`hreysi campaign` fixes all three by moving the **orchestration** (mechanical) into the Go
binary and keeping only the **drafting** (agentic) in spawned agents.

## The split (the same one hreysi runs everywhere)

```
hreysi campaign  (Go — mechanical, deterministic, compiled)
    gate      refuse anything not status: approved
    select    work-list = articles[] not in the brief's drafted: set
    fan out   spawn ONE headless agent per article (bounded concurrency)
    record    on a worker's SUCCESS, append its id to drafted: (crash-safe)
    report    drafted / scaffolded-pending / failed / skipped

  each spawned worker  (agent — generative)
    runs `claude -p` with the writer skill (article-draft | technical-tutorial)
    input: the ArticleBrief rendered as the writer's Step 0
    output: portfolio/articles/NN-<slug>-DRAFT.md
```

The gate and the idempotency guard live in **compiled code**, not a skill's discipline —
which matters because they are what protect article-draft spend. A skill can be prompted
around; a Go `if status != "approved" { return }` cannot.

## Why hreysi spawns agents (and the billing landmine)

hreysi can't draft — drafting is a skill, which only the harness runs. So `hreysi campaign`
shells out to headless Claude Code per article:

```
claude -p "<rendered ArticleBrief as Step-0 input>" \
  --append-system-prompt-file <writer-skill-context> \
  --allowedTools "Read,Write,Edit,Bash,Glob,Grep" \
  --print
```

**BILLING — the landmine, already learned once.** The child MUST bill the Claude Code
subscription (`CLAUDE_CODE_OAUTH_TOKEN`), not metered API. If `ANTHROPIC_API_KEY` /
`ANTHROPIC_AUTH_TOKEN` are in the environment, the Agent SDK bills the API — an early-2026
LifeOS incident was exactly this (every Telegram message a 25-turn API session; see
`LIFEOS/PULSE/modules/telegram.ts:35-43`). So the spawner **deletes those keys from the
child env** before exec. This is non-negotiable and belongs in a test.

## Mechanical guarantees (all in Go)

- **Approved-only gate.** `status != approved` → exit 0, "brief is a draft, approve it
  first." The runner never flips status; approval is the human's token authorization.
- **Idempotent, per-article.** Read `drafted:` from the brief. Draft only ids not in it.
  Append each id **after** its worker succeeds, one at a time (atomic frontmatter rewrite),
  so a crash re-runs only unfinished articles. Re-running a fully-drafted brief spends
  nothing.
- **Bounded fan-out.** `--concurrency N` (default `min(cpu-2, article_count)`). N isolated
  workers at a time; the blowout is bounded, not a thundering herd.
- **Partial failure is contained.** A worker that fails (non-zero exit, no output file, or a
  malformed brief like a `kind: tutorial` with no `capability`) leaves its id **out** of
  `drafted:`, is reported, and does not block the others. A re-run retries only the failures.
- **Tutorial scaffold, not full build.** For `kind: tutorial`: draft the prose worker, then
  mechanically `mkdir sample-repos/<slug>/`, write the TEACH.md stub + `PACKAGE-PENDING.md`.
  The clean-room repo extraction and `learn/step-N` staging stay the (later)
  tutorial-package-builder's job — hreysi flags them, never fakes them.

## Command surface

```
hreysi campaign status [<brief>]     # what's approved / drafted / pending, no side effects
hreysi campaign run <brief>          # gate → fan out → draft the undrafted articles
    --concurrency N                  #   bound the fan-out (default min(cpu-2, count))
    --dry-run                        #   print the work-list + planned NN numbers, spawn nothing
    --only a1,t1                     #   restrict to specific article ids
hreysi campaign run --watch          # poll the Campaigns dir; fire run when a brief flips to approved
```

`--watch` mirrors `hreysi watch`: the human's *approval* is the trigger (they flip
`status: approved` in their editor), and hreysi picks it up — so even the "fire" step leaves
the loop, while approval stays the human decision. `--dry-run` is the cheap way to see what a
run would cost before spending a token.

## The interactive skill stays

`article-runner` (the skill) and `hreysi campaign run` share the **same contract** — the
approved gate and the `drafted:` set. Keep both:

- **skill** — interactive: you're already in a session and say "draft this brief." One
  context, conversational, good for one article or a careful watch.
- **subcommand** — headless/scheduled: bounded parallel fan-out, isolated workers,
  cron-able, no live session. The blowout handler.

They never fight because idempotency is in the brief, not the tool: whichever drafts an
article records it, and the other skips it.

## Where it sits in the weekly rhythm

```
nightly   hreysi mirror                         (Go, cron-safe)
weekly    hreysi digest → digest-to-brief       (digest Go; brief needs an agent — a
                                                  scheduled headless invocation, same spawn
                                                  mechanism as campaign) → brief(draft)
          → notify the human "a brief is ready"
  ── human reviews + flips status: approved ──   (the one human decision)
on approval  hreysi campaign run --watch fires   → article drafts in the portfolio
          → notify "drafts ready for review"
```

The human touches the loop **once** — the approval — which is the only judgment call. Digest
is mechanical; brief-writing and drafting are agent work hreysi fires; the fan-out is
bounded and isolated. Everything else is plumbing.

## Open design questions (resolve at build time)

1. **Writer-skill context injection.** `--append-system-prompt-file` wants a file; the writer
   skills are `SKILL.md`s. Either point at the skill file directly, or render a per-article
   prompt that `Skill()`-invokes the writer. Decide by testing which the harness honors
   headless.
2. **Output capture.** Does the worker write the portfolio file itself (Write tool) and
   hreysi verifies its existence, or does hreysi capture stdout and write? Prefer the former
   (the worker owns its artifact; hreysi checks it landed) — simpler, and it matches how the
   skill already works.
3. **`--watch` cadence + de-dupe.** Poll interval, and how to avoid double-firing a brief
   that's mid-run (a lockfile per brief, or the `drafted:` set as the lock).
4. **Notifications.** Reuse the LifeOS voice/notify channel (`:31337/notify`) or stay silent
   and let the report suffice. Optional.

## Phased plan

- **P1 — `hreysi campaign status` + `run --dry-run`.** Pure Go: parse briefs, gate, compute
  the work-list, print planned NN numbers. No agent spawn. Cheap, testable, immediately
  useful ("what would this cost?"). This is where the gate + idempotency + brief parser get
  written and tested — the trustworthy core, no tokens.
- **P2 — `run` with a single worker.** Spawn one headless agent for one article; verify the
  billing scrub, output capture, and `drafted:` record. Prove the seam end to end on one
  article before fanning out.
- **P3 — bounded fan-out.** `--concurrency`, partial-failure containment, tutorial scaffold.
  The blowout handler proper.
- **P4 — `--watch`.** Approval-triggered firing; the lockfile/de-dupe; optional notify. This
  is what makes the weekly rhythm hands-off past approval.

Each phase ships independently and is useful alone. P1 alone already answers "how do I fire
it and what will it cost" without spending anything.

## Invariants

1. **The gate is compiled, not prompted.** `status != approved` → nothing spawns. The token
   authorization is the human's approval, enforced in Go.
2. **Idempotency lives in the brief.** The `drafted:` set is the shared truth; skill and
   subcommand both honor it; re-runs never re-spend.
3. **Workers bill the subscription.** API-key env stripped before every spawn. Tested.
4. **Partial failure never cascades.** One bad article is reported and retried next run; the
   rest proceed.
5. **hreysi orchestrates; agents generate.** The binary never drafts prose, and a skill never
   owns the gate. Same split as capture/expand, digest/digest-to-brief.
