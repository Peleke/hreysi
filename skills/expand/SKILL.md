---
name: expand
description: Expand today's hreysi commit spine into a narrative buildlog entry using the current session's lived experience. Use at the end of (or during) a work session to turn mechanical commit logs into the story of what happened — the decisions, dead ends, and lessons — emitting a channel-neutral `threads[]` seam that downstream clients (linwheel, article writers, newsletters) project into content.
---

# hreysi expand

Turn the mechanical commit spine hreysi captured into a **narrative** buildlog
entry, using the lived experience in this session — the part git can't see.

Commit metadata (hash, message, files, timestamp) is a skeleton: it says *what*
changed and *when*, never *why* or *how it felt to get there*. That lives in the
session — what was tried, what broke, what was chosen and why. Expansion is the
**join** of the two. The spine is your index and timeline; the session is the story.

You are also the **last point in the pipeline that witnessed the work.** Everything
downstream — post drafters, article writers, campaign sequencers — reads only what
you write. If you don't record it, it did not happen. That is why this skill emits
a machine-readable `threads[]` seam alongside the prose.

## When to run

- **Ambiently**, at the end of a work session (the design target), or
- **Manually**, mid-session for long sessions, or when preparing a demo / content.

## Inputs

1. **The spine** — `buildlog/<today>.md`, specifically the `## Commits` section
   hreysi wrote. Use the hashes, messages, files, and timestamps as the timeline
   you anchor the narrative to. (If unsure which file, pick the entry whose
   commits match the work in this session.)
2. **The lived experience** — THIS conversation and the repo state. What was
   actually attempted, what fought back, the dead ends, the decisions and their
   reasoning, the moment it clicked.

> **Do not invent lived experience.** Draw only on what genuinely happened in the
> session and is visible in the repo. If a commit came from work you didn't
> witness, summarize from the diff and say so — never fabricate a story. The value
> of a buildlog is that it's true.

## The shape of an entry

hreysi creates the file mechanically as `# <date>` + `## Commits`. Expansion
**restructures** it into the form below. The spine is preserved verbatim and moves
to the bottom — it is an index, not the story.

```markdown
---
title: "<a thesis-bearing title — the claim, not the date>"
date: 2026-07-14
project: <repo name>
tags: [<topical>, <topical>]
status: <shipped — what landed, or: in progress / abandoned>
threads:
  - id: <kebab-slug>
    thesis: "<the single claim this thread makes>"
    evidence: "<what in the work actually proves it>"
    section: "## <the heading in this file this thread draws from>"
    tags: [<topical-slug>, <topical-slug>]   # the CLUSTERING keys — see below
    angle: field_note | contrarian | synthesizer | demystifier | curious_cat
    scale: post | article | both
    fit: <1-10>
    facts:
      - "<entity/version/name that MUST survive downstream generation verbatim>"
---

# <the thesis-bearing title>

## <topical section>          ← session-specific. one per real thread of work.
## <topical section>          ← each is a candidate source chunk for a client.

## The Journey
## Improvements
## Roadmap / next

## Commits                    ← hreysi owns this. Preserve VERBATIM. Keep it last.
```

## What to write

### Frontmatter

This is **routing metadata for machines**, not decoration. Downstream clients read
it instead of re-mining your prose.

- **`title`** — a claim, not a date. `# 2026-07-14` tells a campaign sequencer
  nothing; *"The config that was never read"* gives it a thesis to build an arc on.
- **`project`**, **`tags`**, **`status`** — what shipped, and where it belongs.
- **`threads[]`** — see below. This is the seam.

### `threads[]` — the seam (the most important thing you write)

One entry per *distinct idea* the session produced. A thread is **not a topic** —
it is a **claim with evidence behind it**. "I worked on MCP config" is a topic and
is worthless. "Two MCP servers had never once loaded because the config file was
never read" is a thread.

Keep it **channel-neutral**. Do not write LinkedIn posts, article outlines, or any
channel's vocabulary here. Describe the *idea*; let each client project it. linwheel
is the first client, not the only one.

- **`thesis`** — one sentence. The claim. If you can't state it in one sentence,
  it isn't a thread yet.
- **`evidence`** — what in the actual work proves it. This is the anti-fabrication
  anchor: a downstream writer who can't find evidence must not embellish.
- **`tags[]`** — the **clustering keys**: 2-4 topical slugs naming *what this thread is
  about* (`silent-failure`, `git-hooks`, `provenance`). The weekly digest converges
  threads that share a tag into campaign themes, so pick tags a *future, unrelated*
  thread could also carry. This is different from `facts[]` (below): a tag is a topic
  two threads might share; a fact is a specific name that must survive verbatim. Reuse
  tags across sessions deliberately — a tag only clusters if it recurs. Distinct from
  the entry-level `tags` in the header, which summarize the whole day; these are
  per-thread and finer.
- **`angle`** — the genre that fits the idea. Portable; each client maps it.
- **`scale`** — `post` (a single beat), `article` (needs long-form: a thought piece
  and/or its paired technical tutorial), or `both`.
- **`fit`** — 1-10, honest. Most work is a 3. Reserve 8+ for genuine surprise —
  something that changed *your* mind. Downstream triage trusts this number.
- **`facts[]`** — **the fact-lock.** Any named entity, model, version, or product
  whose exact wording must survive generation. Generators routinely "helpfully"
  normalize unfamiliar names into familiar ones. You saw the work; you are the only
  one who can pin these. List them.

### Topical sections

One `##` section per real thread of work, titled for *what it is*, not generically.
`## The gap we closed`, `## The admin write path` — each becomes a **source chunk**
a client can draw a single piece of content from. This is what makes sequencing
possible: a campaign sequencer needs discrete beats, not one undifferentiated blob.

### `## The Journey`

A first-person narrative of the session, anchored to the commit timeline:

- What you set out to do, and why it mattered.
- **The friction** — what failed, what failed *silently*, the wrong turns. The
  dead ends are the most reshareable part; keep them, don't sand them off.
- The decisions and the *why*, especially where you chose between real alternatives.
- The moment(s) it clicked.

Specific beats general. Honest beats flattering. This is the raw material
downstream tools reshape into content — vague narrative makes vague content.

### `## Improvements`

Atomic, reusable lessons — takeaways that transfer to the next project. Bullet
them under H3 category headers so downstream distillation stays compatible:

- `### Architectural`
- `### Process`
- `### Tooling`
- `### Gotchas`

Omit categories with nothing to say. One lesson per bullet, stated as a rule, not
a retelling.

### `## Roadmap / next`

What this session *opened*. Known gaps, deferred work, the thing you'd do next.
Downstream this is what makes a campaign feel like a series instead of a list.

## Contract

- The `buildlog/` directory is the seam. You only *read* `## Commits`; you write
  everything else. Mechanical capture (hreysi) and narrative (expansion) meet in
  the file and never overwrite each other.
- **`## Commits` is preserved verbatim and kept LAST.** hreysi's capture appends
  new commits to the end of that section; keeping it final is what makes a
  same-day commit land correctly after you've expanded.
- **Idempotent:** if the sections already exist, refine them to cover newer
  commits rather than duplicating. Merge new threads into `threads[]`; don't
  restate existing ones.
- **Stamp when done:** after writing the narrative, record that this commit has
  been expanded so the ambient trigger doesn't re-fire:
  ```sh
  git rev-parse HEAD > buildlog/.hreysi/last_expanded
  ```

## Gotchas

- **A thread without evidence is a lie waiting to be published.** Downstream
  writers cannot check your work; they will confidently expand whatever you assert.
  If the session didn't prove it, don't thread it.
- **`fit` inflation poisons the pipeline.** If everything is an 8, triage is dead
  and the corpus generates noise. Most sessions produce no 8s. That is normal.
- **Don't write channel content here.** The moment `threads[]` contains a hook, a
  CTA, or a post, the corpus is coupled to one client and the next client inherits
  LinkedIn's shape. Describe the idea. Nothing else.
