# Tutorial Package — design

> **Status:** spec / design riff. The mistake-ledger *port* (§5) is defined as a concrete
> seam to build now; the ledger *solution* behind it ships later. Everything else is the
> agreed shape of the long-form tutorial track.
>
> Sits downstream of `campaign-brief-spec.md` → an `ArticleBrief` of `kind: tutorial`
> expands into the package below. Read that spec first for how a brief is produced.

## Why a "package," not an article

A thought piece argues a thesis; a tutorial **builds a capability**. So a tutorial is not
one artifact — it is three, shipped together:

```
tutorial ArticleBrief (approved)
   ├── prose          the written tutorial            (brunnr `technical-tutorial`)
   ├── sample repo    a clean-room, runnable example  (extracted, NOT the source)
   └── TEACH.md       an agent-led interactive walkthrough of the sample repo
```

The prose is what brunnr already writes. The sample repo and TEACH.md are the new pieces,
and they are the reason a tutorial gets *proposed* only when a real, extractable capability
exists (per the brief's pairing rule) — not reflexively.

## 1. When a tutorial is proposed

`digest-to-brief` proposes a `tutorial` ArticleBrief for a cluster **only if a generalized,
runnable sample repo can be cleanly extracted** from its threads — a minimal teachable
version of the idea, not the product. Thought-piece-eligible ≠ tutorial-eligible. The bar
is: *could a reader rebuild this concept from a small standalone repo?* If not, it stays a
thought piece. This is decided per brief; the human confirms.

## 2. The sample repo — clean-room extraction

**The sample repo teaches the CONCEPT, never ships the source.** This is a hard rule, and
it is what makes the track safe for private and NSFW codebases (e.g. `hush`, the flagship):

```
hush thread: "a guard matched too loosely and failed silent"
        │  extract the IDEA, generalize, strip all product specifics
        ▼
sample-repos/loose-match-silent-failure/
   minimal · runnable · generic · zero source from the origin repo · zero NSFW
```

- Built fresh as a standalone repo. Never a copy or excerpt of the origin.
- Generalized to the concept: the reader should learn the *pattern*, decoupled from where
  it happened.
- Provenance is one-way and forgetful — the sample repo does not reference or link back to
  the private source. (Same discipline as `mirror`'s clobber guard and the daemon's
  security filter: the private side never leaks downstream.)
- If the concept already lives in a **public** repo (e.g. hreysi itself), the tutorial may
  point at that instead of extracting a new one. `digest-to-brief` flags which, based on
  whether a public home exists; the human confirms.

## 3. TEACH.md — the instructor/student oscillation

TEACH.md is an **agent-followable** doc that drives an *interactive* session over the
sample repo. It is not a second README, and not the tutorial prose re-stated — it is a
script for an agent to teach a human by walking them through the repo and then **handing
off**.

**Lineage, and the delta.** It inherits the pedagogy the house already uses:

- aegir `lesson-generator` — the **DO → NOTICE → CODE → NAME** fractal loop.
- aegir `outline-writer` — the **Starting State Rule** (state exactly what the learner has
  coming in; non-negotiable).
- brunnr `technical-tutorial` — SPIN arc + that same loop, for the prose.

The **new** thing is the *oscillation*: TEACH does not run the loop *at* the reader, it
alternates instructor and student turns.

```
TEACH.md turn structure (per concept primitive):
  INSTRUCTOR   orient — point at the entry, state the goal, invoke Starting State
  INSTRUCTOR   DO / NOTICE — have them run it, observe the real behavior
  ── HANDOFF ──▶ STUDENT does: fill the gap, perform the task, read + answer
  INSTRUCTOR   checkpoint — "explain why that happened" / verify the student's work
  (miss? → §5 records it. hit? → advance to the next primitive.)
```

The handoff is the point. A static tutorial tells; TEACH makes the student *do*, then
checks. The agent is the instructor; the human is the student; control passes back and
forth.

## 4. The canonical path is staged in git

When a concept has steps the student completes, the sample repo encodes them as git
history — the repo *is* the lesson state machine:

```
main                     the complete, canonical solution (the answer key)
learn/step-1             staged INCOMPLETE — a gap where the student writes code,
                         with the instructions inline (a TODO the TEACH turn points at)
learn/step-2             step-1 solved, step-2's gap now open
…                        each branch: prior steps done, the current one waiting
```

- The student checks out `learn/step-N`, TEACH walks them to the gap, they fill it.
- Verification is a diff/test against the canonical version (or the next `learn/step-N+1`
  as the reference solution).
- `main` is never handed over first — it is the answer key, revealed after (or used by the
  instructor to check).

So the "canonical path" (the expected completion sequence) is not prose — it is real,
checkoutable, runnable repository state. A student can always run what they have, and always
compare against what it should become.

## 5. The mistake-ledger PORT (build this; solution later)

When a student **misses** at a checkpoint, that miss is signal — eventually the most
valuable signal in the system. But the ledger's *analysis* is a later build. What we build
now is the **port**: a stable seam a solution drops into, plus a lightweight default so
nothing is lost in the meantime.

### The record contract (stable)

Every miss (and every resolution) is recorded as:

```
MistakeRecord {
  ask       the question / task the student was given
  response  what the student answered or did
  gap       where it fell short of the checkpoint (the delta)
  # resolution, appended when they get there:
  hints     the ordered hints/insights offered after the miss
  resolved  the hint/insight that actually landed (nullable until it does)
  concept   the primitive being taught (for later aggregation)
}
```

That `resolved` field is the fulcrum the whole thing is for — *which* insight moved the
student from miss to understanding. **We do not compute it yet.** We just make sure the
record can hold it, so the later analysis has ground truth.

### The pluggable seam

TEACH, at a miss, does exactly this:

```
on miss:
  if a correction/error-log skill is available → hand it the MistakeRecord (it owns storage + analysis)
  else                                          → append the MistakeRecord to a local ledger (the lightweight default)
```

- **The check is capability-detection, not a hard dependency.** TEACH asks "is there a
  correction skill?" and uses it if present. This is the port: a solution plugs in by
  *existing*, not by TEACH being rewritten.
- **The lightweight default** is a flat append to `sample-repos/<name>/.teach/mistakes.jsonl`
  — one `MistakeRecord` per line. Zero analysis, never blocks the lesson, never lost.
- The eventual solution (the "which hint was the fulcrum" analyzer) replaces the default by
  registering as the correction skill. Same contract, richer behavior. **No TEACH change
  when it arrives** — that is what makes this a port.

> Deliberately deferred: hint-efficacy analysis, cross-student aggregation, the
> insight-that-unlocks-understanding fulcrum. The port guarantees the data for all of it is
> captured from day one, in a shape the analyzer can consume, whether or not the analyzer
> exists yet.

## How it connects back

```
digest → digest-to-brief → Campaign Brief (draft)
                               articles[]:
                                 - kind: thought   → article-draft
                                 - kind: tutorial  → THIS PACKAGE, iff extractable
   ── human approves (the token gate — nothing below runs until then) ──
   runner:
     thought  → article-draft            → portfolio/articles/NN-*-DRAFT.md
     tutorial → technical-tutorial (prose)
              + clean-room sample repo (extract/generalize)
              + TEACH.md (oscillation script)
              + learn/step-N git staging
              + mistake-ledger port wired (correction-skill check → lightweight default)
```

The tutorial `ArticleBrief` therefore carries a bit more than a thought one — enough for the
runner to build the package: the capability to teach, the extraction target (new clean-room
vs existing public repo), and the step breakdown for the canonical path. Those fields extend
`campaign-brief-spec.md`'s ArticleBrief and should be specced there when the runner is built.

## Invariants

1. **The sample repo never contains origin source.** Concept, generalized, clean-room. This
   is what makes the track safe for private/NSFW codebases.
2. **TEACH oscillates.** Instructor walks, then hands off; the student does, then is checked.
   A TEACH.md that only tells is a README, not a lesson.
3. **The canonical path is git state, not prose.** `main` is the answer key; `learn/step-N`
   stages the gaps.
4. **The mistake ledger is a port, not a feature (yet).** Build the seam + the record
   contract + a lightweight default. The analyzer plugs in later by registering as the
   correction skill, with no change to TEACH.
5. **Nothing token-heavy runs before approval.** The package builds only after the human
   flips the brief to approved. Article drafts are the cost; the gate protects them.
