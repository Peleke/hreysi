---
name: digest-to-brief
description: Turn a hreysi Digest Report into a Campaign Brief — the connector between mechanical clustering and content generation. Reads the week's clustered corpus (via `hreysi digest`), NAMES each campaign candidate, and writes a review-ready Campaign Brief (status draft) with article briefs and orbiting beats. Manual and opt-in. Names clusters; never invents membership. Nothing generates, nothing publishes — the brief is inert until a human approves it.
---

# hreysi digest-to-brief

Turn the mechanical **Digest Report** (`hreysi digest`) into a **Campaign Brief** — a
review-ready plan for the week's content, written to the vault as `status: draft`. This is
the agentic *naming* half of digest: the report already decided WHAT converges; you decide
what to CALL it and how to shape it. This is a **manual, opt-in** step. Content generation
is a choice.

> Nothing here generates articles or posts, and nothing publishes. You write a **brief**.
> A human reviews it and flips it to `approved`; only then does the (token-heavy) drafting
> run. The brief is the review gate that protects article-draft spend.

## The one hard rule

**You NAME clusters; you never change membership.** The Digest Report's clusters were
chosen by set intersection on explicit tags — mechanically, reproducibly. You may phrase a
theme, draft its thesis, sequence its beats, and decide article vs post shaping. You may
**not** merge two clusters an LLM "feels" are related, split one, or add a thread the report
didn't cluster. If you think the clustering is wrong, that is a tagging problem to fix
upstream (in `expand`), not something to override here. Overriding membership reintroduces
exactly the confabulation the mechanical split exists to prevent.

## Inputs

1. **The Digest Report.** Run `hreysi digest` (optionally `--since`/`--until`) in a repo
   with a configured vault. It prints campaign candidates (≥2-thread clusters), standalone
   posts, a tag-drift surface, and a coverage note. Use it verbatim as your selection.
2. **The threads behind it.** For each thread the report references, read its full entry in
   `<vault>/Buildlog/YYYY-MM-DD-<repo>.md` — the `thesis`, `evidence`, `facts`, and the
   `## The Journey` prose. This is where claims, evidence, and war stories come from.

If `hreysi digest` reports no vault or no threads, stop and say so — there is nothing to
brief.

## Steps

1. **Run digest.** Capture the report. Note the coverage warning verbatim — it goes in the
   brief so a thin week is never mistaken for a complete one.
2. **Per campaign candidate (a cluster):**
   - **Name the theme** and write a one-sentence **thesis** — the claim its threads share.
     Ground it in the threads' own theses; do not reach past them.
   - **Build article brief(s).** A cluster with `scale: article`/`both` threads anchors an
     `ArticleBrief` of `kind: thought`. Its `editorial_map` is one entry per claim, each
     carrying the thread's `evidence` **copied, not paraphrased into fiction** — a claim
     with no evidence is dropped. `war_stories` come from `## The Journey`. `facts` = the
     union of the cluster's thread facts.
   - **Propose a tutorial only if the bar is met** (see below).
   - **Sequence beats** that orbit the article: one per remaining thread, ordered by arc
     role (opener → standout → gut-check → conclusion), each pointing at its `thread` id.
3. **Per standalone post:** a single beat (or a single-article brief if it's substantial and
   `article`-scaled). Most singles are one post.
4. **Assemble the Campaign Brief** per `docs/campaign-brief-spec.md` and write it to
   `<vault>/Campaigns/YYYY-Www.md` with **`status: draft`**.
5. **Report.** Tell the user the brief is in the vault for review, name its theme(s), and
   state plainly: nothing drafts or publishes until they flip it to `approved`.

## The tutorial bar

Propose an `ArticleBrief` of `kind: tutorial` (paired to a thought piece via `pairs_with`)
**only when a generalized, runnable sample repo can be cleanly extracted** from the
cluster's threads — a minimal teachable version of the concept, never the source. If the
work didn't produce a rebuildable capability, it's a thought piece, full stop. See
`docs/tutorial-package-spec.md` for what the package entails; here you only *propose* it and
record the extraction intent (new clean-room vs an existing public repo). The human confirms
at approval. Do not pair reflexively — most clusters are thought-only.

## Contract

- **Draft only.** Always `status: draft`. Never flip to `approved` yourself; that is the
  human's token-gate.
- **Evidence or silence.** Every `editorial_map` claim carries evidence copied from a
  thread. No evidence → drop the claim. This is the anti-fabrication invariant, end to end.
- **Channel-neutral.** The brief describes ideas and arc roles, not LinkedIn posts. linwheel
  and the article pipeline project it; you don't write their output.
- **Carry the coverage note.** Put the report's un-threaded-repo warning in the brief's
  `## Notes`, so an approver knows what digest could not see.
- **Fact-lock survives.** `facts[]` is the union of the drawn threads' facts; downstream
  clients assert it. Never normalize a name the threads pinned.

## Gotchas

- **Naming is not membership.** If the theme you'd write requires threads from two different
  clusters, you're confabulating — write two briefs (or fix the tags upstream), don't merge.
- **A brief with no article-eligible cluster is still valid** — it's a week of standalone
  posts. Don't manufacture an article to make it feel like a campaign.
- **Don't over-pair tutorials.** The extractable-sample-repo bar is high on purpose. When in
  doubt, thought piece only.
- **The coverage note is not decoration.** A campaign that would obviously be stronger with a
  repo digest couldn't see is worth flagging to the human in your report, not silently
  shipping half of it.

## Examples

- "build this week's campaign brief" → run `hreysi digest`, read the clustered threads, write
  `<vault>/Campaigns/2026-W29.md` (draft) with a thought-piece article for the top cluster,
  beats orbiting it, the coverage warning in Notes. Tell the user it's ready to review.
- "brief the silent-failure cluster as a thought+tutorial pair" → only if a clean-room demo
  repo is extractable; otherwise brief it as thought-only and say why.
