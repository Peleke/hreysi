# Handoff: "Your commits are your content" (the pipeline article)

> A brief, not a draft. Third article seeded from the 2026-07-06 session, alongside
> the Homebrew and Golang handoffs. This one is the cross-cutting piece: the whole
> linwheel + hreysi + LifeOS pipeline as a lived build. Writing happens elsewhere.

## One-line thesis

The content you should be posting already exists as a byproduct of your engineering.
The missing piece is a pipeline that captures it, narrates it, and reshapes it into
your voice, with one human approval at the end. We assembled and proved that pipeline
in a single session, and the buildlog of that session became the demo's own input.

## Audience & promise

- **Audience:** engineers who ship constantly and post nothing; people evaluating
  linwheel; the beta user.
- **Promise:** they leave understanding a concrete pipeline they can install, not a
  motivational take on "personal brand." Every claim is backed by a real run.

## Structure

1. **The problem, plainly.** You do the work. The record of it evaporates. "Make
   content" becomes a separate chore, so you skip it. The bottleneck is not ideas,
   it is the reshape-and-post tax plus not sounding like a bot.
2. **The pipeline in one diagram.** commit → capture → narrate → reshape → approve.
   For each stage: what it is, and whether it is automatic, weekly-manual, or human.
3. **The three parts.**
   - **linwheel** reshapes text into voice-matched LinkedIn posts. The voice profile
     is what stops the output being generic slop (the extractor rejects off-voice
     samples). Live in production: publishing, carousels, scheduling.
   - **hreysi** installs a git post-commit hook that writes each commit to a dated
     buildlog file; a skill narrates those commits from the session's own context.
   - **LifeOS / PAI** makes capture and narration ambient by wiring a SessionEnd hook,
     because hreysi is built from the same primitives (skills, hooks, a capture dir).
4. **The proof, with numbers.** A plain journal entry scored 7/10 for LinkedIn fit and
   came back as a real contrarian post. A buildlog entry scored 8/10 (structured and
   already about the work) and reshaped into a field-note post. Both runs were live.
5. **The design principle.** Keep the layers decoupled through a directory: producers
   only write it, consumers only read it. Content generation stays opt-in and manual;
   nothing auto-generates or auto-publishes. Explain why that restraint matters:
   trust, voice, and not flooding a feed on autopilot.
6. **The meta-move.** The session that built the pipeline was captured by the pipeline.
   The article's own examples come from that buildlog. Show the loop closing on itself.
7. **Close.** Start with one post a week, pulled from what you already shipped.

## Concrete anchors (all real, 2026-07-06)

- linwheel: `analyze` returned fit 7 (journal) and 8 (buildlog); `reshape` produced
  real posts in both cases. Voice extractor kept 5 real samples, rejected 1 at 0.97.
- hreysi: shipped to v1.1.0 (brew + curl + releases), 6 commands, expand + reshape
  skills, ambient SessionEnd hook. The whole tool dogfooded its own build.
- The walkthrough deck (private) is the visual spine: linwheel → the turn → hreysi →
  LifeOS.

## Tone & constraints

- House voice, engineer to engineer. Mechanism first. No em dashes in prose (bragi
  #10). No "load-bearing," no puffery, no negative parallelisms. Show real output.

## Sources
- `hreysi/buildlog/2026-07-06.md`, the walkthrough deck, `hreysi/docs/lifeos-integration.md`,
  the linwheel onboarding notes (`linwheel/docs/onboarding-walkthrough-notes.md`).

---

## Backlog note (the parenthetical, made explicit)

This is now **three articles from one session**: Homebrew how-to, Golang in the age of
agents, and this pipeline piece. They are a **narrative series**, and right now they
live as scattered handoffs, which is exactly the tracking gap worth closing.

**Need:** a single backlog for narrative series and campaigns, so seeded articles and
their status (idea → drafted → scheduled → published) are visible in one place instead
of decaying in per-repo docs. Portfolio already has the seeds of this
(`portfolio/linwheel-queue-2026-07.md`, `linwheel-campaign-mode-handoff.md`).
Recommend consolidating the three handoffs above into that queue as a tracked series.
