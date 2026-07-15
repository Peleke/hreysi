# `hreysi digest` — design

> **Status:** spec. Not yet implemented. Consumes the mirrored corpus; produces a
> Campaign Brief (see `campaign-brief-spec.md`). Read that first.

## What digest is

The weekly step that turns a period of the mirrored corpus into **one Campaign Brief**.
It answers a single question: *of everything captured this period, what converges into
a theme worth a campaign, and what's just a post?*

It has two halves, and keeping them apart is the whole design:

| Half | Who | Trust model |
|---|---|---|
| **Selection** — what converges | mechanical (Go) | deterministic, testable, inspectable |
| **Naming** — what to call it | agentic (skill) | LLM, but only *phrases* what selection already decided |

This is the same split hreysi already runs everywhere: **the binary writes mechanical
truth; a skill adds the story.** Capture ⟶ expand. Cluster ⟶ name. The LLM never
decides membership, only wording — so a confabulated "theme" can't enter the brief,
because the themes are chosen by set intersection before the LLM sees anything.

## Selection: convergence is set intersection, not noticing

A **thread** (from `threads[]` in a mirrored entry) carries explicit `tags` and
`facts`. digest clusters threads by **shared keys**:

```
cluster(period):
  for each pair of threads (a, b) in the period:
    if a.keys ∩ b.keys ≠ ∅:  they converge
  a connected component of ≥2 threads = a candidate CAMPAIGN theme
  a thread in no component = a single POST
```

That is the entire convergence rule. No LLM reads free text to "find the throughline."
If two threads share the tag `silent-failure`, they converge; if they don't, they
don't. This is what makes the output defensible and is the direct consequence of the
reach-back decision: **without a concept-graph substrate, convergence is only what's
explicitly shared — anything softer is vibes over text search.**

Worked example — the corpus entry that produced this spec (`2026-07-14-hreysi.md`):

```
doctor-false-green        tags: silent-failure, doctor   ┐
verdict-ignored-checks    tags: silent-failure, doctor   ├─ share silent-failure + doctor
fixtures-encoded-the-bug  tags: silent-failure, testing  ┘   → 3-thread component → CAMPAIGN
init-wont-repair          tags: doctor                   ── shares doctor with above → joins it
mirror-clobber-guard      tags: mirror, provenance       ── no shared key → single POST
```

Selection alone, no model, gives: one campaign theme (4 threads on silent-failure/doctor)
and one standalone post (mirror). The skill then *names* the campaign — but it could not
have invented it, and it cannot drop the mirror thread into it, because the keys don't
intersect.

### `fit` is a prior, not truth

`expand` self-reports `fit: 1-10` per thread. digest **does not rank on it directly** —
a session that inflates every thread to 8 would poison triage. `fit` is a prior;
digest re-weights on corpus evidence:

- does the thread have real `evidence` (not empty)?
- does it carry a `war_story`-grade anecdote in its section?
- **how many other threads does it converge with?** (cross-thread degree is the
  strongest signal — a claim with three witnesses outranks a lonely self-reported 9)

Convergence degree beats self-reported fit. That ordering is the anti-inflation lever.

### Single article vs multi-article, derived

Straight from the cluster sizes, not stipulated:

- **component of 1** → single-article brief + orbiting beats (a normal week)
- **component of ≥2** → multi-article campaign; the shared claim *is* the thesis
- **thread with `scale: article` or `both`** → eligible to anchor an ArticleBrief;
  `scale: post` threads only ever become beats
- **tutorial pairing gates on reproducibility** — a thought piece gets a `tutorial`
  twin only if some thread in its cluster produced a rebuildable capability (a repo,
  CLI, config). digest does not pair reflexively.

## Coverage: name the blind spot, never hide it

digest reads the mirrored corpus, which only contains repos that captured **and**
expanded **and** mirrored. A week where only one repo did all three looks, to a naive
digest, like a complete week. That is the exact silent-success failure this tool's own
origin story is about — so digest refuses it.

Every brief carries a **coverage note**:

```
⚠ coverage — 2026-W29
  captured + threaded this period:  hreysi (5 threads)
  session activity, NOT threaded:   LifeOS, linwheel
  → the cross-project version of "silent-failure" is NOT visible to digest.
     run capture in those repos to make it minable.
```

How digest knows what it's missing (best-effort, mechanical): compare the set of repos
with mirrored entries in the window against a signal of where work happened —
`~/Documents/Projects/*/buildlog/*.md` with recent un-mirrored/un-expanded entries. It
can only report gaps it can see; it says so, and never implies completeness it can't
verify. **No silent truncation** — if digest bounds anything (top-N clusters, a thread
cap), it logs what it dropped.

## Output

A `CampaignBrief` (`status: draft`) at `<vault>/Campaigns/YYYY-Www.md`, per
`campaign-brief-spec.md`:

- one `ArticleBrief` per article-eligible cluster (kind by reproducibility)
- `beats[]` ordered by arc role, orbiting the anchors
- `facts[]` = union of the drawn threads' facts
- the coverage note in the `## Notes` body
- **inert until a human flips `approved`** — content generation stays a choice

digest generates nothing downstream and calls no client. It writes a file. linwheel's
campaign mode and the article pipeline read it later, on their own triggers.

## The split, concretely

```
hreysi digest            (Go, deterministic)
  reads   <vault>/Buildlog/*.md for the period
  parses  frontmatter threads[] (schema-specific, zero-dep line parser)
  emits   a Digest Report: clusters + coverage + fit-adjusted ranking
              — inspectable, no LLM, reproducible, TESTABLE like mirror's guard

digest skill             (agentic)
  reads   the Digest Report + the thread bodies
  writes  the Campaign Brief — NAMES each cluster, drafts its thesis sentence,
          sequences beats, gates tutorial pairing
  cannot  change cluster membership or invent a theme; selection is upstream and fixed
```

Putting clustering in Go (not the skill) is deliberate: convergence is the trustworthy
core, so it must be deterministic and testable — an LLM asked to "intersect the tags"
will eventually drift. The skill gets handed the clusters and may only phrase them.

## Scheduling

Per the pipeline's ambient/deliberate split:

| When | Step | Ambient? |
|---|---|---|
| SessionEnd | `expand` (+ `threads[]`) | ✅ — the transcript exists only here |
| nightly | `hreysi mirror` | ✅ — mechanical, nothing to invent |
| **weekly** | **`hreysi digest` → Campaign Brief (`draft`)** | ✅ — reads corpus, needs no transcript |
| human | flip `draft` → `approved` | 👤 — generation stays a choice |
| client trigger | linwheel / article pipeline consume the approved brief | later |

`expand` **cannot** run on a cron — it needs the session transcript ("the part git
can't see") and would fabricate the narrative without it, poisoning the corpus that
every downstream writer trusts. `mirror` and `digest` are corpus operations and are
cron-safe. The weekly job produces an **inert draft**; nothing generates or publishes
until a human approves.

## Deferred (needs a graph substrate)

- **Reach-back.** A thesis grounded in this week recruiting older threads as evidence.
  This is graph traversal: extract concepts from threads → project each thread into
  concept space → find overlaps → explore the neighborhood. Over flat text it is just
  vibes, so it waits for the substrate. Until then, convergence is **in-window only**.
- **No backfill.** The 34 legacy entries have no `threads[]` and stay mute to digest.
  `newsletter` reads their prose; digest reads threads. Clean split, by decision.

## Invariants

1. **Selection is mechanical.** An LLM never decides what converges — only what to call
   it. A theme absent from the clusters cannot appear in the brief.
2. **Convergence is explicit-key intersection, in-window.** No cross-week reach-back, no
   free-text theming, until a graph substrate exists.
3. **`fit` is a prior.** Rank on evidence and convergence degree, not the self-report.
4. **Coverage is always stated.** A thin week is labeled thin. No silent completeness.
5. **The brief is inert.** `status: draft` until a human approves. digest publishes
   nothing and calls no client.
