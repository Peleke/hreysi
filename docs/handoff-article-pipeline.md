# Handoff — the article pipeline

> **For the agent wiring the long-form path.**
> Read `campaign-brief-spec.md` first. That is the contract.
>
> **The good news: you are not building writers. They exist.** This is an integration
> job — hand the existing brunnr skills the inputs they already ask for.

## Your input

A **Campaign Brief** at `<vault>/Campaigns/YYYY-Www.md`. You consume `articles[]` and
`facts[]`. You **ignore** `beats[]` — those are linwheel's.

Each `ArticleBrief` is one long-form piece. `kind` selects the writer:

| `kind` | Writer skill | Contract |
|---|---|---|
| `thought` | `brunnr/skills/article-draft` | *argues a thesis* |
| `tutorial` | `brunnr/skills/technical-tutorial` | *builds a capability* |

`pairs_with` links a thought piece to its tutorial twin. That pattern is already in the
corpus whether or not it was named: `portfolio/articles/02-non-determinism-canard.md`
(thesis) and `03-actually-testing-llms.md` (practical guide, worked example) **are**
such a pair.

## The chain

```
ArticleBrief
   │
   ├── kind: thought    ──▶ article-draft         ┐
   └── kind: tutorial   ──▶ technical-tutorial    ┘  produce a DRAFT
                                    │
                                    ▼
                            visual-pass                scan the draft, emit a
                                    │                  prioritized injection manifest
                                    ▼
        inline-svg-architecture-diagrams / Manim MCP / ComfyUI MCP / Mermaid
                                    │                  render the assets
                                    ▼
                    engagement-pass  ·  scaffold-pass   final passes
                                    │
                                    ▼
                        portfolio/articles/NN-<slug>-DRAFT.md
```

Every box after `ArticleBrief` **already exists** in `brunnr/skills/`. Your job is the
arrows.

## Why the brief has the fields it has

The two writers refuse to draft from an outline alone. They each declare a `Step 0:
Inputs`, and the Article Brief exists to satisfy it. Do not "simplify" the brief — you
will just move the elicitation pass downstream.

**`article-draft` Step 0 wants:**

1. **Editorial map** — what it claims, *and what evidence supports each claim*
   → the brief's `editorial_map[]` (`claim` + `evidence` + `thread`)
2. **Visual inventory** → **do not supply this.** See below.
3. **War stories** — *specific anecdotes, incidents, real data — not hypotheticals*
   → the brief's `war_stories[]`, drawn from `## The Journey` in the mirrored corpus,
   which hreysi's `expand` skill explicitly forbids fabricating
4. **Voice target** → `voice`

**`technical-tutorial` Step 0 wants something different** (*"articles argue a thesis;
tutorials build a capability"*):

1. **Topic brief** — what the reader will **DO**, not understand → `capability`
2. **Prerequisite inventory** → `prerequisites[]`
3. **Companion artifact type** — notebook / repo / CLI / none → `companion_artifact`
4. **Voice target** → `voice`
5. **War story inventory** → `war_stories[]`

## Visuals: do not put them in the brief

hreysi captures commits and prose. It has no screenshots and no diagrams, and it must
not pretend otherwise.

That is **not a gap** — it is the design. `visual-pass` is *"the analysis layer between
content generation and visual production"*: it scans a **finished draft**, produces a
prioritized injection manifest, and delegates rendering to
`inline-svg-architecture-diagrams`, Manim MCP, ComfyUI MCP, or Mermaid.

So visuals are decided **after** there is prose to look at, by a skill that can see the
prose. If `article-draft` asks for a visual inventory up front, let it run its
elicitation pass, or hand it the manifest `visual-pass` produces on a first pass — but
**never** synthesize a visual inventory into the brief. hreysi would be inventing it.

## Hard rules

- **Refuse to draft from `status: draft`.** A brief is inert until a human approves it.
- **`facts[]` survive verbatim.** Assert post-hoc. A writer will normalize an
  unfamiliar name into a familiar one if you let it.
- **Evidence or silence.** A `claim` with no `evidence` is not draftable. Drop it or go
  back to the corpus. This is the anti-fabrication invariant and the whole reason
  `threads[]` carries evidence at all.
- **Output is a DRAFT**, into `portfolio/articles/`, following the existing naming
  (`NN-<slug>-DRAFT.md`) and the series structure in `portfolio/articles/README.md`.
  Nothing publishes.
- **Do not write to the vault.** hreysi owns the corpus.

## Open questions for the human

1. **Which skill tree is canonical** — `brunnr/skills/` or `hunter/skills/`? They are
   near-identical mirrors, and a third copy is stranded in a hunter worktree
   (`hunter/.claude/worktrees/agent-a72e1ff8/brunnr-work/skills/`). Pick one before
   building against it.
2. `outline-writer` is scoped to **notebook modules** (course content), not articles.
   Confirm it is out of this path.
3. `linwheel-content-engine` and `linwheel-source-optimizer` exist in the same tree and
   overlap linwheel's campaign mode. Decide who owns what before both get built.
