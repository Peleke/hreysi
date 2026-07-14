# Handoff — linwheel campaign mode

> **For the agent building campaign mode in linwheel.**
> Read `campaign-brief-spec.md` first. That is the contract. This document says
> only what campaign mode has to do with it.

## Your input

A **Campaign Brief** — a markdown file with YAML frontmatter at
`<vault>/Campaigns/YYYY-Www.md`, produced by `hreysi digest`.

You consume exactly two things from it:

- **`beats[]`** — the ordered post sequence. Each beat points at either a `thread`
  (an idea, defined in a mirrored buildlog entry) or an `article` (a long-form anchor
  another pipeline is writing).
- **`facts[]`** — entity names that must survive your generation verbatim.

You **ignore** `articles[]`. That belongs to the article pipeline. A beat with
`article: a1` means *"this beat is the article's slot in the sequence"* — you are not
writing the article, you are sequencing around it.

## What campaign mode has to fix

`linwheel_reshape` is stateless and prompt-based. Per
`portfolio/linwheel-campaign-mode-handoff.md` (the by-hand reference run, 2026-07-06),
it has no memory, no arc, and no feedback loop, so it:

1. **silently renamed a named entity** (`Fable` -> "Anthropic's new Claude 4 model"),
2. **re-injected em dashes** every run, against house style, because the active voice
   profile *endorses* them in its own description and samples, and
3. **optimized each post in isolation** — five good posts, no throughline.

All three were fixed **by hand** that run. Campaign mode is the durable fix.

> The campaign's own thesis, turned on the tool that publishes it: the work isn't in
> the generation call, it's in the clarity of the spec you hand it.

## The three things that make it work

**1. Generate against the arc, not in isolation.**
For each beat, reshape its source against **the shared contract + its `arcRole` + its
neighboring beats**. A beat that doesn't know what came before it cannot carry a
throughline. This is the entire value of campaign mode; per-post polish is not.

**2. Enforce a contract, don't prompt for one.**
"No em dashes" belongs in a **deterministic validator**, not a hope in a prompt. Run a
validation pass that **rejects contract violations before a draft is ever saved**:

```
contract: {
  factsLocked: [...],        # from the brief's facts[]. Assert post-hoc, every beat.
  forbid: ["em dash", "renaming named entities"],
  voiceProfileId: ...,
}
```

Fact-lock is not optional. A generator will "helpfully" normalize any name it doesn't
recognize. The brief pins those names because expansion was the last stage that
actually witnessed the work.

**3. Fix the voice profile, not the drafts.**
The `peleke-linkedin` profile *endorses* em dashes in its description and samples.
Cleaning them downstream every run is treating the symptom. Strip at the profile level
(re-extract), then validate.

## The data model

The reference doc maps this near 1:1 onto the existing **storyboard generator/editor**:

| Storyboard | Campaign mode |
|---|---|
| Storyboard | Campaign (theme + thesis + cadence) |
| Scene | Beat (day, angle, source, arcRole) |
| Shot | Post (the drafted text) |
| Companion clips | Carousel + boost comment |
| Continuity / style bible | **Voice + fact contract** (enforced every generation) |

Adapt it. Don't invent a new one.

## Hard rules

- **Refuse to generate from `status: draft`.** A brief is inert until a human approves
  it. Content generation is a choice.
- **Never publish.** Drafts land in the linwheel dashboard. Approval and scheduling
  stay in the editor, where they already work (149 posts and counting).
- **Map `angle` -> `postType` on your side.** The corpus is channel-neutral on purpose;
  do not push LinkedIn vocabulary back into it.
- **Do not write to the vault.** hreysi owns the corpus. It never reads your output.

## Known bugs to fix or route around (from the reference run)

- `linwheel_post_carousel_companion` inlines the rendered PDF **and every slide PNG as
  base64** into the tool response (~1–1.5M chars per call). Should return `carouselUrl`
  + text.
- Carousel slicer **splits on periods**: "Opus 4.8" becomes two slides ("Opus 4" /
  "8 and through Fable"). Needs decimal-aware chunking.
- Auto-caption/CTA is generic on every carousel ("Swipe through the key takeaways").
  Campaign mode should generate these **from the beat**.

## Done means

- A brief with N beats produces N drafts that read as **one sequence**, not N good posts.
- Every name in `facts[]` survives verbatim, asserted post-hoc.
- Zero em dashes, enforced by a check, not a prompt.
- Nothing was published.

---

## Second by-hand reference run — 2026-07-14

Two articles (`agent-pack`, `dossier-generator`) reshaped into 5 posts each. Confirms the
original three failures and adds **two the first run did not surface**. Root causes traced
to source; cite these instead of rediscovering them.

### Confirmed

- **Em-dash re-injection.** Still happening, same cause. The `peleke-linkedin` profile
  description literally read *"Comfortable with ellipses and em dashes for rhythm."*
  **Now fixed at the profile level** (prescription 3 above): sample containing em dashes
  scrubbed, one outright-slop sample deleted, description rewritten to *ban* dashes. The
  deterministic validator is still required. A profile edit is not a contract.

### New: the reshape bandit silently fans out and ignores the active profile

One call with 5 angles returned **15 posts** (3 voice variants per angle), attributed to
profiles that were **not active** (`Technical Architect`, `Retrospective`, and a synthetic
`voice:unshaped` arm). An identical call minutes earlier returned exactly 5.

- `buildVoiceCandidates` (`linwheel/src/lib/voice/bandit.ts:65-80`) builds arms from
  **every** profile, with **no `isActive` filter**, plus a synthetic unshaped arm.
  `isActive` is honored **only** in the fallback path (`buildFallback`, `bandit.ts:142-163`).
- `k` is **hardcoded to 3** at `linwheel/src/app/api/agent/reshape/route.ts:205`.
- The nondeterminism: `qortexFetch` (`linwheel/src/lib/qortex.ts:65-97`) has a 10s timeout
  and **swallows every error, returning `null`**, which the caller treats as "not
  configured" and silently falls back to the active profile. Same input, different output,
  no error surfaced.
- Deeper, already documented in `linwheel/docs/LEARNING-LAYER-CRITIQUE.md` (Gap 1): `select`
  reads the `"default"` qortex context partition while `observeReward` writes a unique
  per-post context, so **the posteriors never update**. Selection is effectively
  uniform-random across all profiles, not learned.

**Implication for campaign mode:** you cannot assume one beat produces one post. Either add
a `voiceMode: "bandit" | "active"` param (~20-30 lines: MCP schema, route body, one
early-return in `bandit.ts`), or treat generation as over-producing and make **best-of
selection a required stage** of the pipeline. The contract validator should also assert the
returned post count matches the beat count.

### New: there is no post-delete tool, so over-generation is unrecoverable

The 10 unwanted drafts could not be removed. They had to be marked `[ALT, NOT SELECTED]`
in-place. With the bandit over-producing 3x, every campaign run permanently litters the
dashboard (currently 156 posts, mostly abandoned drafts).

**This is trivial to fix and blocks nothing else:**
1. Add a `DELETE` handler to `linwheel/src/app/api/agent/posts/[postId]/route.ts` (only
   `GET`/`PATCH` today). Copy the working user-facing one at
   `src/app/api/posts/[postId]/route.ts:171-208`, swapping `requireAuth` for
   `requireAgentAuth` — the same substitution already made for GET/PATCH in that file.
2. Add `linwheel/mcp/src/tools/post-delete.ts` (~15 lines; mirror `post-carousel-delete.ts`).
3. Register in `mcp/src/tools/index.ts`.

No schema change. `scheduledComments` already cascades on post delete (`schema.ts:416`);
`imageIntents` does not, so delete it first, as the existing handler does.

### New: scheduled comments never fire if you publish manually

Boost comments attach fine (`status: "pending"`), and the mechanism is genuinely wired for
the **auto-publish cron** path. But the dashboard's **"Publish Now" button never calls
`activateScheduledComments()`**:

- `linwheel/src/app/api/posts/[postId]/publish-linkedin/route.ts:132-140` and
  `.../publish-org/route.ts:149-158` both set `linkedinPostUrn`/`linkedinPublishedAt` and
  skip activation, unlike the cron path which calls it at
  `src/app/api/cron/auto-publish/route.ts:324-326` and `:378-380`.
- `getCommentsToFire` only selects `status: "scheduled"` (`comment-scheduler.ts:71`), so a
  manually-published post's comments are stranded at `pending` **forever**.
- Fix: one call, 2-4 lines, in each of the two manual-publish routes.
- Also note `/api/cron/auto-publish` is **not** in `vercel.json` (Hobby-plan cron limit); it
  is triggered externally via cron-job.org. Whether that trigger is live cannot be verified
  from the repo.

**Until this is fixed:** keep CTA links **in the post body**. Do not rely on the boost
comment to carry the only link.
