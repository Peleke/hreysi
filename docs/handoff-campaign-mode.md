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
