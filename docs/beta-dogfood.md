# Beta: running hreysi + LifeOS on ourselves

> Private runbook. The plan is to dogfood the full stack — LifeOS with hreysi
> wired in — on our own machine before/while onboarding beta users, so we prove
> the workflow on ourselves first.

## 0. Stand up LifeOS

```sh
# Back up first — LifeOS installs into ~/.claude
cp -r ~/.claude ~/.claude-backup-$(date +%Y%m%d)

# Install from Miessler's repo (LifeOS / Personal_AI_Infrastructure)
#   https://github.com/danielmiessler/Personal_AI_Infrastructure
# Follow its README installer. Your customizations live under USER/ and are
# preserved across upgrades.
```

Verify LifeOS is live: `~/.claude/skills/`, `~/.claude/hooks/`, and its agentic
routing are present.

## 1. Wire hreysi in

```sh
brew install Peleke/tap/hreysi
hreysi version

# Register the expand skill globally (LifeOS layout, capital S)
hreysi skills --global

# Opt in to content generation for ourselves (we want the linwheel loop)
hreysi skills --global --linwheel

# Turn on capture + ambient expansion for each repo we work in
cd ~/Documents/Projects/<repo>
hreysi init --ambient --linwheel
hreysi doctor
```

Do this for the repos we actually push to (start with hreysi itself, cadence,
buildlog-template).

## 1b. Prove it on ourselves

Run the real loop for a few days and check each stage produced output:

1. **Capture** — commit normally; confirm `buildlog/YYYY-MM-DD.md` grows a
   `## Commits` block per commit. `hreysi doctor` stays green.
2. **Narrative** — after a Claude Code session, confirm `## The Journey` +
   `## Improvements` were written (SessionEnd hook). If a `buildlog/.hreysi/
   pending_expansion` marker appears instead, the headless path wasn't reachable —
   run the `expand` skill manually and note it (this is the best-effort edge to
   harden).
3. **Digest** — at end of week, run the `reshape` skill; confirm drafts appear in
   the linwheel dashboard, triaged (only fit >= 7), with sane angles.
4. **Human loop** — review/approve/schedule a couple of the drafts in linwheel.

Log what breaks in a buildlog entry (dogfood the tool with the tool). The known
soft spot to watch: the SessionEnd headless expansion — if it's flaky, that's the
next thing to harden before beta users hit it.

## Alongside beta users

Once the loop holds for us for a week, hand `docs/lifeos-integration.md` to a beta
user's agent and have it run the same steps. Their first-week experience should
match section "The daily rhythm" there. Collect friction; feed it back here.
