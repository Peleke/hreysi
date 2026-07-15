// Command hreysi is an ambient buildlog capture tool: every git commit is
// appended to a dated markdown journal, with no ceremony and nothing to
// remember. The journal directory is the whole product — point anything at it.
package main

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Peleke/hreysi/internal/ambient"
	"github.com/Peleke/hreysi/internal/campaign"
	"github.com/Peleke/hreysi/internal/capture"
	"github.com/Peleke/hreysi/internal/digest"
	"github.com/Peleke/hreysi/internal/gitx"
	"github.com/Peleke/hreysi/internal/mirror"
	"github.com/Peleke/hreysi/internal/scaffold"
	"github.com/Peleke/hreysi/internal/skillpack"
	"github.com/Peleke/hreysi/internal/watch"
)

// version is overridden at release time via -ldflags "-X main.version=...".
var version = "dev"

// skillFS carries the skills that ship with hreysi (dropped into
// .claude/skills/ by `hreysi init`).
//
//go:embed skills
var skillFS embed.FS

// optSkillFS carries opt-in skills (e.g. the linwheel `reshape` weekly digest),
// installed only when the user asks (`--linwheel`). Content generation is a
// choice, never on by default.
//
//go:embed optskills
var optSkillFS embed.FS

// hookFS carries the ambient expansion hook script (dropped by
// `hreysi init --ambient`).
//
//go:embed hooks
var hookFS embed.FS

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	switch os.Args[1] {
	case "init":
		os.Exit(cmdInit())
	case "capture":
		os.Exit(cmdCapture())
	case "doctor":
		os.Exit(cmdDoctor())
	case "watch":
		os.Exit(cmdWatch())
	case "skills":
		os.Exit(cmdSkills())
	case "mirror":
		os.Exit(cmdMirror())
	case "digest":
		os.Exit(cmdDigest())
	case "campaign":
		os.Exit(cmdCampaign())
	case "version", "--version", "-v":
		fmt.Printf("hreysi %s\n", version)
	case "help", "--help", "-h":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "hreysi: unknown command %q\n\n", os.Args[1])
		usage()
		os.Exit(2)
	}
}

func cmdInit() int {
	noSkill := false
	linwheel := false
	var ambientEvents []string
	for _, a := range os.Args[2:] {
		switch a {
		case "--no-skill":
			noSkill = true
		case "--linwheel": // opt in to the reshape weekly-digest skill
			linwheel = true
		case "--ambient": // default trigger: expand at end of session
			ambientEvents = []string{"SessionEnd"}
		case "--ambient-stop": // also expand on Stop (long / ultramarathon sessions)
			ambientEvents = []string{"SessionEnd", "Stop"}
		}
	}

	cwd, _ := os.Getwd()
	exe, err := os.Executable()
	if err != nil || exe == "" {
		exe = "hreysi" // fall back to PATH resolution
	}
	res, err := scaffold.Init(cwd, exe)
	if err != nil {
		fmt.Fprintf(os.Stderr, "hreysi: %v\n", err)
		return 1
	}
	fmt.Printf("hreysi initialized in %s\n", res.Root)
	fmt.Printf("  journal:  %s/\n", res.EntryDir)
	fmt.Printf("  hook:     %s (%s)\n", res.HookPath, res.HookAction)
	if res.Warning != "" {
		fmt.Printf("  note:     %s detected — run `hreysi doctor` to confirm capture fires\n", res.Warning)
	}

	if !noSkill {
		dest := filepath.Join(res.Root, ".claude", "skills")
		if written, serr := skillpack.Install(dest, skillFS); serr != nil {
			fmt.Fprintf(os.Stderr, "hreysi: skill install warning: %v\n", serr)
		} else if len(written) > 0 {
			fmt.Println("  skill:    .claude/skills/expand — run it to narrate your day into the entry")
		}
		if linwheel {
			if _, serr := skillpack.Install(dest, optSkillFS); serr != nil {
				fmt.Fprintf(os.Stderr, "hreysi: linwheel skill warning: %v\n", serr)
			} else {
				fmt.Println("  skill:    .claude/skills/reshape — weekly digest → linwheel drafts (opt-in)")
			}
		}
	}

	if len(ambientEvents) > 0 {
		if err := ambient.Install(res.Root, hookFS, ambientEvents); err != nil {
			fmt.Fprintf(os.Stderr, "hreysi: ambient hook warning: %v\n", err)
		} else {
			fmt.Printf("  ambient:  expansion hook wired for %s (best-effort, non-fatal)\n",
				strings.Join(ambientEvents, "+"))
		}
	}

	fmt.Println("  every commit now lands in today's entry — nothing to remember.")
	return 0
}

func cmdCapture() int {
	cwd, _ := os.Getwd()
	root, err := gitx.RepoRoot(cwd)
	if err != nil {
		fmt.Fprintln(os.Stderr, "hreysi: not a git repository")
		return 1
	}
	out, err := capture.Once(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "hreysi: %v\n", err)
		return 1
	}
	switch out.Action {
	case "skipped":
		fmt.Printf("hreysi: %s already captured\n", out.Hash)
	case "amended":
		fmt.Printf("hreysi: amended %s → %s\n", out.Hash, out.Path)
	default:
		fmt.Printf("hreysi: captured %s → %s\n", out.Hash, out.Path)
	}
	return 0
}

func cmdSkills() int {
	global, linwheel := false, false
	for _, a := range os.Args[2:] {
		switch a {
		case "--global":
			global = true
		case "--linwheel":
			linwheel = true
		}
	}

	var dest string
	if global {
		// Claude Code / LifeOS read ~/.claude/skills (lowercase). The docs say
		// "Skills" but the installer and the harness both use lowercase, and on
		// case-sensitive filesystems only lowercase is read.
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "hreysi: %v\n", err)
			return 1
		}
		dest = filepath.Join(home, ".claude", "skills")
	} else {
		cwd, _ := os.Getwd()
		if root, err := gitx.RepoRoot(cwd); err == nil {
			cwd = root
		}
		dest = filepath.Join(cwd, ".claude", "skills")
	}

	written, err := skillpack.Install(dest, skillFS)
	if err != nil {
		fmt.Fprintf(os.Stderr, "hreysi: %v\n", err)
		return 1
	}
	if linwheel {
		more, lerr := skillpack.Install(dest, optSkillFS)
		if lerr != nil {
			fmt.Fprintf(os.Stderr, "hreysi: %v\n", lerr)
			return 1
		}
		written = append(written, more...)
	}
	for _, w := range written {
		fmt.Printf("hreysi: installed %s\n", w)
	}
	return 0
}

func cmdWatch() int {
	cwd, _ := os.Getwd()
	root, err := gitx.RepoRoot(cwd)
	if err != nil {
		fmt.Fprintln(os.Stderr, "hreysi: not a git repository")
		return 1
	}
	fmt.Printf("hreysi: watching %s — every commit captured (any client), Ctrl-C to stop\n", root)
	err = watch.Run(root, time.Second, func() {
		if out, e := capture.Once(root); e == nil && out.Action != "skipped" {
			fmt.Printf("hreysi: %s %s\n", out.Action, out.Hash)
		}
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "hreysi: %v\n", err)
		return 1
	}
	return 0
}

// cmdCampaign is the P1 orchestrator surface: `status` and `run --dry-run`. It gates on
// approval and plans the article fan-out WITHOUT spawning agents or spending tokens — the
// agentic drafting (P2+) lands later. See docs/campaign-run-spec.md.
func cmdCampaign() int {
	cwd, _ := os.Getwd()
	vault := mirror.VaultDir(cwd)
	if vault == "" {
		fmt.Println("hreysi: no vault configured — campaign reads briefs from the vault.")
		fmt.Println("  set one with:  hreysi mirror --vault <path>")
		return 0
	}
	campaignsDir := filepath.Join(vault, "Campaigns")
	portfolioDir := filepath.Join(filepath.Dir(cwd), "portfolio", "articles")

	args := os.Args[2:]
	sub := "status"
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		sub = args[0]
		args = args[1:]
	}
	dryRun := false
	var briefPath string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--dry-run":
			dryRun = true
		default:
			if !strings.HasPrefix(args[i], "-") {
				briefPath = args[i]
			}
		}
	}

	// Resolve the brief: an explicit path, else the newest in Campaigns/.
	if briefPath == "" {
		briefs, err := campaign.FindBriefs(campaignsDir)
		if err != nil || len(briefs) == 0 {
			fmt.Printf("hreysi: no campaign briefs in %s\n", campaignsDir)
			fmt.Println("  generate one with the digest-to-brief skill after `hreysi digest`.")
			return 0
		}
		briefPath = briefs[0]
	} else if !filepath.IsAbs(briefPath) {
		briefPath = filepath.Join(campaignsDir, briefPath)
	}

	b, ok, err := campaign.ParseBrief(briefPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "hreysi: %v\n", err)
		return 1
	}
	if !ok {
		fmt.Printf("hreysi: %s is not a campaign brief\n", filepath.Base(briefPath))
		return 1
	}
	plan := campaign.BuildPlan(b, portfolioDir)

	fmt.Printf("hreysi campaign — %s  (%s)\n", b.Period, filepath.Base(briefPath))
	fmt.Printf("  theme:  %s\n", b.Theme)
	fmt.Printf("  status: %s\n", b.Status)
	if len(plan.Skipped) > 0 {
		fmt.Printf("  drafted already: %s\n", strings.Join(plan.Skipped, ", "))
	}

	switch sub {
	case "status":
		fmt.Println("  articles:")
		for _, a := range b.Articles {
			mark := "•"
			if campaign.InDrafted(b, a.ID) {
				mark = "✓"
			}
			fmt.Printf("    %s %s (%s) — %s\n", mark, a.ID, a.Kind, truncate(a.Summary, 60))
		}
		if plan.Blocked != "" {
			fmt.Printf("\n  ⚠ %s\n", plan.Blocked)
		} else if len(plan.WorkList) == 0 {
			fmt.Println("\n  all articles drafted — a run would spend nothing.")
		} else {
			fmt.Printf("\n  approved — a run would draft %d article(s), starting at %02d-\n", len(plan.WorkList), plan.NextNum)
		}
		return 0

	case "run":
		if plan.Blocked != "" {
			fmt.Printf("\n  ⚠ %s\n", plan.Blocked)
			return 1 // the gate: a non-approved brief is a failed run, loudly
		}
		if len(plan.WorkList) == 0 {
			fmt.Println("\n  nothing to draft — all articles are in drafted:. No tokens spent.")
			return 0
		}
		fmt.Printf("\n  would draft %d article(s):\n", len(plan.WorkList))
		n := plan.NextNum
		for _, a := range plan.WorkList {
			writer := "article-draft"
			if a.Kind == "tutorial" {
				writer = "technical-tutorial (+ package scaffold)"
			}
			fmt.Printf("    %02d- %s (%s) → %s\n", n, a.ID, a.Kind, writer)
			n++
		}
		if dryRun {
			fmt.Println("\n  --dry-run: no agents spawned, no tokens spent.")
			return 0
		}
		// P1 stops here — the agentic fan-out (spawning claude -p per article) is P2+.
		fmt.Println("\n  (agent fan-out is not built yet — this is P1: gate + plan only.)")
		fmt.Println("  fire the drafts now via the `article-runner` skill in a Claude Code session,")
		fmt.Println("  or wait for `hreysi campaign run` P2 (headless per-article workers).")
		return 0

	default:
		fmt.Fprintf(os.Stderr, "hreysi campaign: unknown subcommand %q (want: status | run)\n", sub)
		return 2
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}

// cmdDigest reads the mirrored corpus for a window and prints a Digest Report:
// tag-clustered campaign candidates, standalone posts, tag drift, and a coverage note.
// It is the mechanical half — it selects what converges; a naming skill writes the
// Campaign Brief from it. Window defaults to the last 7 days; --since/--until override.
func cmdDigest() int {
	cwd, _ := os.Getwd()
	vault := mirror.VaultDir(cwd)
	if vault == "" {
		fmt.Println("hreysi: no vault configured — digest reads the mirrored corpus.")
		fmt.Println("  set one with:  hreysi mirror --vault <path>   (then mirror your entries)")
		return 0
	}

	now := time.Now()
	since := now.AddDate(0, 0, -7).Format("2006-01-02")
	until := now.Format("2006-01-02")
	args := os.Args[2:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--since":
			if i+1 < len(args) {
				since = args[i+1]
				i++
			}
		case "--until":
			if i+1 < len(args) {
				until = args[i+1]
				i++
			}
		}
	}

	buildlogDir := filepath.Join(vault, mirror.VaultSubdir)
	threads, repos, err := digest.ParseCorpus(buildlogDir, since, until)
	if err != nil {
		fmt.Fprintf(os.Stderr, "hreysi: %v\n", err)
		return 1
	}

	// Coverage looks at sibling repos under the parent of the current repo.
	projectsDir := filepath.Dir(cwd)
	cov := digest.DetectCoverage(projectsDir, since, until, repos)

	rep := digest.Build(fmt.Sprintf("%s … %s", since, until), threads, cov)
	fmt.Print(rep.Markdown())
	return 0
}

// cmdMirror copies expanded entries into an Obsidian vault, so a week of work across
// every repo can be read as one corpus. Optional: with no vault configured it says so
// and exits clean — hreysi must never require anyone's note-taking setup.
func cmdMirror() int {
	cwd, _ := os.Getwd()

	// `hreysi mirror --vault <path>` records the vault for this repo, then mirrors.
	args := os.Args[2:]
	for i := 0; i < len(args); i++ {
		if args[i] == "--vault" {
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "hreysi: --vault needs a path")
				return 2
			}
			if err := mirror.SetVault(cwd, args[i+1]); err != nil {
				fmt.Fprintf(os.Stderr, "hreysi: %v\n", err)
				return 1
			}
			i++
		}
	}

	res, err := mirror.Run(cwd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "hreysi: %v\n", err)
		return 1
	}
	if res.VaultDir == "" {
		fmt.Println("hreysi: no vault configured — mirroring is off.")
		fmt.Println("  set one with:  hreysi mirror --vault ~/path/to/ObsidianVault")
		fmt.Println("  or export HREYSI_VAULT_DIR")
		return 0
	}

	fmt.Printf("hreysi mirror — %s\n", filepath.Join(res.VaultDir, mirror.VaultSubdir))
	for _, d := range res.Mirrored {
		fmt.Printf("  ✓ %s\n", filepath.Base(d))
	}
	for _, s := range res.Skipped {
		fmt.Printf("  · %s\n", s)
	}
	for _, r := range res.Refused {
		fmt.Printf("  ✗ %s\n", r)
	}
	fmt.Println()
	switch {
	case len(res.Mirrored) == 0 && len(res.Refused) == 0:
		fmt.Println("nothing to mirror — expand an entry first.")
	case len(res.Refused) > 0:
		// Loud, but not fatal: refusing is mirror working correctly, not failing.
		fmt.Printf("%d mirrored, %d left untouched (not hreysi's files).\n", len(res.Mirrored), len(res.Refused))
	default:
		fmt.Printf("%d mirrored — the corpus is current.\n", len(res.Mirrored))
	}
	return 0
}

func cmdDoctor() int {
	cwd, _ := os.Getwd()
	rep, err := scaffold.Check(cwd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "hreysi: %v\n", err)
		return 1
	}
	glyph := map[string]string{"ok": "✓", "warn": "!", "fail": "✗"}
	fmt.Printf("hreysi doctor — %s\n", rep.Root)
	fmt.Printf("  hooks dir: %s\n", rep.HooksDir)
	for _, c := range rep.Results {
		line := fmt.Sprintf("  %s %s", glyph[c.Level], c.Name)
		if c.Detail != "" {
			line += " — " + c.Detail
		}
		fmt.Println(line)
	}
	if rep.Healthy {
		fmt.Println("\ncapture is live — every commit will be journaled.")
		return 0
	}
	fmt.Println("\ncapture will NOT fire reliably. Run `hreysi init` to (re)install the hook.")
	return 1
}

func usage() {
	fmt.Print(`hreysi — ambient buildlog capture

Every git commit, appended to a dated journal. No ceremony.

USAGE:
  hreysi init       Scaffold buildlog/ and install the post-commit capture hook
                    (--ambient wires a SessionEnd expansion hook; --ambient-stop adds Stop;
                     --linwheel adds the opt-in weekly reshape→linwheel digest skill)
  hreysi capture    Append HEAD to today's entry (run by the hook; also manual)
  hreysi watch      Watch the reflog and capture every commit — any client, can't-miss
  hreysi doctor     Check that capture is actually wired and will fire
  hreysi skills     Install the bundled skills (--global → ~/.claude/skills for LifeOS/PAI;
                    --linwheel → include the opt-in reshape digest skill)
  hreysi digest     Cluster the week's mirrored corpus into campaign candidates +
                    standalone posts (--since/--until to set the window)
  hreysi mirror     Copy expanded entries into an Obsidian vault, so a week of work
                    across every repo reads as ONE corpus (--vault <path> to configure).
                    Optional and one-way: never overwrites a file hreysi didn't write.
  hreysi version    Print version
  hreysi help       Show this help

The buildlog/ directory is the whole product: dated markdown, one timestamped
commit block each. Point a narrative-expansion skill, a content pipeline, or a
learning loop at it — hreysi only ever writes; consumers only ever read.
`)
}
