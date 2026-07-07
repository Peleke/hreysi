---
name: reshape
description: Weekly digest — scan the week's expanded buildlog entries, triage the post-worthy ones, and reshape them into LinkedIn drafts via linwheel. Manual and opt-in. Use at end of week (or on demand) to turn buildlogs into draft posts in your linwheel dashboard. Nothing auto-publishes.
---

# hreysi reshape — weekly digest → linwheel

Turn the week's **narrated** buildlog entries into LinkedIn draft posts, dropped
into your linwheel dashboard for review. This is a **manual, opt-in** step — a
deliberate weekly fire, never an ambient trigger. Content generation is a choice.

> Nothing here publishes. Drafts land in linwheel; you review, approve, and
> schedule them there, where the editor already works.

## When to run

End of week, or on demand. If this skill isn't installed, hreysi simply never
generates content — that's the opt-out.

## Inputs

- **The week's narrative.** `buildlog/YYYY-MM-DD.md` for roughly the last 7 days —
  the `## The Journey` and `## Improvements` sections. Use the *story*, not the raw
  `## Commits` spine; the narrative is what reshapes into good content. Skip days
  with no narrative (run the `expand` skill for those first if you want them in).
- **linwheel MCP tools** (`linwheel_analyze`, `linwheel_reshape`). If they aren't
  available, stop and tell the user to set up the linwheel MCP first.

## Steps

1. **Scan.** Collect the narrative text from the last ~7 days of entries.
2. **Triage.** For each distinct thread/day, call `linwheel_analyze(text,
   context="buildlog entry: <date>")`. Keep threads with `linkedinFit.score >= 7`;
   drop the rest — a weekly digest is a highlight reel, not everything. Record the
   suggested angles.
3. **Reshape.** For each kept thread, call `linwheel_reshape(text, angles=<top 1–2
   suggested angles>, saveDrafts=true)` so the drafts land in the dashboard.
4. **Report.** Tell the user which threads became drafts, which angles, and that the
   drafts are waiting in linwheel for review / approve / schedule.

## Notes

- `saveDrafts=true` is intentional here: this is a manual weekly fire, so landing
  drafts in the dashboard *is* the goal. There is still no automatic publishing.
- **Opt-out** = don't install this skill (`hreysi init` without `--linwheel`).
  Nothing else in hreysi references linwheel, so opting out is total.
