# Campaign Brief — the publication seam

> **Status:** spec. Not yet implemented. This document is the contract between
> hreysi (which produces briefs) and every client that consumes them.
>
> **Audience:** agents. If you are building a client, read this, not hreysi's source.

## What this is

hreysi captures work and narrates it. It does **not** write content. The **Campaign
Brief** is the artifact it hands off — a plain markdown file with YAML frontmatter,
written into the Obsidian vault, that any client can read.

The rule that shapes everything below:

> **hreysi never knows its clients exist.** It writes a brief to the seam and stops.
> linwheel is the first client, not the only one. No LinkedIn vocabulary, no MCP
> calls, no client-specific fields ever enter the corpus.

If you find yourself wanting to add a field because *your* client needs it, ask
whether the *next* client would want it too. If not, derive it on your side.

## The hierarchy

```
CampaignBrief                      one period (default: a week) of work
  ├── ArticleBrief   (1..N)        the long-form ANCHORS
  │     kind: thought | tutorial   ← different input contracts; see below
  │     pairs_with: <id>           ← a thought piece and its tutorial twin
  └── beats[]        (0..N)        the short-form posts that ORBIT the anchors
```

A Campaign Brief with **one** ArticleBrief and a handful of beats is a normal week.
A Campaign Brief with **several** ArticleBriefs is a multi-article campaign. Both are
the same artifact; nothing special-cases the single-article case.

**The article is the anchor; beats orbit it.** This is not a stylistic preference —
it is what the reference run actually did (see
`portfolio/linwheel-campaign-mode-handoff.md`: beat 2 *was* the article, five builder
posts orbited it, cadence W·F·M·W·F).

## Where briefs live

```
<vault>/Campaigns/YYYY-Www.md          e.g. Campaigns/2026-W29.md
```

Source material is the mirrored corpus:

```
<vault>/Buildlog/YYYY-MM-DD-<repo>.md  written by `hreysi mirror`
```

## The schema

```yaml
---
type: campaign-brief
period: 2026-W29
theme: "The silent failure"
thesis: "The tools built to detect silent failure fail silently themselves"
cadence: "W·F·M·W·F"          # advisory; the client owns scheduling
status: draft                  # draft | approved — nothing generates from a draft

# Entities whose exact wording MUST survive generation. Generators normalize
# unfamiliar names into familiar ones ("Fable" -> "Claude 4"). Pinned at the point
# of witness, because expansion is the last stage that actually saw the work.
facts:
  - "hreysi 1.1.1"
  - "Fable is Anthropic's newest model"

# ── The long-form anchors ───────────────────────────────────────────────────
articles:
  - id: a1
    kind: thought                     # -> brunnr `article-draft`
    thesis: "A tool that reports success while failing is worse than one that crashes"
    # article-draft Step 0 requires claims WITH the evidence for each. This is that.
    editorial_map:
      - claim: "doctor reported capture live over its own failing check"
        evidence: "Healthy = exists && wired && executable — omitted the target check"
        thread: doctor-false-green
      - claim: "init could not repair what doctor told you to run init for"
        evidence: "idempotency keyed on the marker comment alone"
        thread: init-wont-repair
    war_stories:                      # REAL incidents. Never hypotheticals.
      - "hreysi's own repo pointed at /tmp/hreysi-build/hreysi, deleted weeks earlier"
    voice: <voice-profile-id>
    # NOTE: no `visuals:` field. Visual inventory is a POST-draft pass — see
    # brunnr `visual-pass`, which scans the finished draft and emits an injection
    # manifest. hreysi captures no visuals and must not pretend to.

  - id: t1
    kind: tutorial                    # -> brunnr `technical-tutorial`
    pairs_with: a1                    # the thought/tutorial pair
    # tutorials have a DIFFERENT contract: "articles argue a thesis; tutorials
    # build a capability." State what the reader will DO, not understand.
    capability: "Write a git hook whose failure is impossible to miss"
    prerequisites: [git, shell]
    companion_artifact: repo          # notebook | repo | cli | none
    war_stories:
      - "`|| true` swallowed the ENOENT for weeks; every commit 'succeeded'"
    voice: <voice-profile-id>

# ── The short-form beats that orbit them ────────────────────────────────────
beats:
  - order: 1
    arcRole: opener
    angle: field_note                 # portable genre; the client maps it
    thread: dead-mcp-config           # source: a thread id...
    source: Buildlog/2026-07-14-hreysi.md
  - order: 2
    arcRole: the-anchor
    article: a1                       # ...or the article itself
  - order: 3
    arcRole: the-standout
    angle: field_note
    thread: doctor-false-green
    source: Buildlog/2026-07-14-hreysi.md
---

## Notes

Free prose. Anything a human wants the clients to know that the schema can't hold.
```

## Field semantics that matter

**`thread`** — references a `threads[].id` in a mirrored buildlog entry. That entry
carries the thread's `thesis`, `evidence`, `angle`, `scale`, and `fit`. **A beat or
article never restates thread content; it points at it.** One source of truth.

**`angle` / `arcRole`** — deliberately channel-neutral. `angle` is the genre of the
*idea* (`field_note`, `contrarian`, `synthesizer`, `demystifier`, `curious_cat`);
`arcRole` is its position in the narrative (`opener`, `the-anchor`, `the-standout`,
`gut-check`, `conclusion`). A client maps these into its own vocabulary. linwheel maps
`angle` -> `postType`. The next client maps it to something else. **Neither mapping
belongs in the corpus.**

**`status: draft`** — a brief is inert until a human sets `approved`. Clients MUST
refuse to generate from a `draft`. Content generation stays a choice; capture and
narration are the only ambient parts of this system.

**`facts[]`** — the union of every `threads[].facts` the brief draws on. Clients
should assert these survive their own generation, post-hoc.

## Invariants

1. **A brief is a file, not an API call.** Any client can read it. hreysi has no
   client-specific code paths and no client-specific dependencies.
2. **Nothing generates from a `draft`.** Human sets `approved`.
3. **Nothing publishes.** Clients produce drafts. Approval and scheduling live in the
   client's own editor (linwheel's dashboard, a PR in `portfolio`, …).
4. **Evidence or silence.** An article with claims and no `evidence` is not draftable;
   both brunnr writers refuse to draft from an outline alone, and they are right to.
5. **The corpus is append-only from hreysi's side.** hreysi writes briefs and mirrored
   entries. It never reads a client's output back.

## Who builds what

| Piece | Owner | Status |
|---|---|---|
| `threads[]` in expanded entries | hreysi | ✅ shipped (PR #3) |
| `hreysi mirror` — repo -> vault | hreysi | **to build** |
| `hreysi digest` — corpus -> Campaign Brief | hreysi | **to build** |
| Campaign mode — consumes `beats[]` | linwheel | see `handoff-campaign-mode.md` |
| Article pipeline — consumes `articles[]` | brunnr skills | see `handoff-article-pipeline.md` |
