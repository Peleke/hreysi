---
name: expand
description: Expand today's hreysi commit spine into a narrative buildlog entry using the current session's lived experience. Use at the end of (or during) a work session to turn mechanical commit logs into the story of what happened — the decisions, dead ends, and lessons — ready to reshape into content (e.g. linwheel).
---

# hreysi expand

Turn the mechanical commit spine hreysi captured into a **narrative** buildlog
entry, using the lived experience in this session — the part git can't see.

Commit metadata (hash, message, files, timestamp) is a skeleton: it says *what*
changed and *when*, never *why* or *how it felt to get there*. That lives in the
session — what was tried, what broke, what was chosen and why. Expansion is the
**join** of the two. The spine is your index and timeline; the session is the story.

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

## What to write

Edit `buildlog/<today>.md`. **Never modify the `## Commits` section — hreysi owns
it.** Add or refine these two sections in the same file:

### `## The Journey`

A first-person narrative of the session, anchored to the commit timeline:

- What you set out to do, and why it mattered.
- **The friction** — what failed, what failed *silently*, the wrong turns. The
  dead ends are the most reshareable part; keep them, don't sand them off.
- The decisions and the *why*, especially where you chose between real alternatives.
- The moment(s) it clicked.

Specific beats general. Honest beats flattering. This is the raw material
downstream tools reshape into posts — vague narrative makes vague content.

### `## Improvements`

Atomic, reusable lessons — takeaways that transfer to the next project. Bullet
them under H3 category headers so downstream distillation stays compatible:

- `### Architectural`
- `### Process`
- `### Tooling`
- `### Gotchas`

Omit categories with nothing to say. One lesson per bullet, stated as a rule, not
a retelling.

## Format

The `format` argument selects the render target (default `buildlog`):

- **`buildlog`** (shipped) — the narrative entry described above.
- Other formats (`linwheel`, `portfolio`, …) are provided by additional hreysi
  skills. Same join `(spine ⋈ lived experience)`, different projection.

## Contract

- The `buildlog/` directory is the seam. You only *read* `## Commits`; you only
  *write* `## The Journey` / `## Improvements`. This keeps mechanical capture
  (hreysi) and narrative (expansion) decoupled — they meet in the file, never
  overwrite each other.
- **Idempotent:** if the sections already exist, refine them to cover newer
  commits rather than duplicating.
