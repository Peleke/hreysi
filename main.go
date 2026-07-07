// Command hreysi is an ambient buildlog capture tool: every git commit is
// appended to a dated markdown journal, with no ceremony and nothing to
// remember. The journal directory is the whole product — point anything at it.
package main

import (
	"embed"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Peleke/hreysi/internal/ambient"
	"github.com/Peleke/hreysi/internal/capture"
	"github.com/Peleke/hreysi/internal/gitx"
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
	var ambientEvents []string
	for _, a := range os.Args[2:] {
		switch a {
		case "--no-skill":
			noSkill = true
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
		if written, serr := skillpack.Install(res.Root, skillFS); serr != nil {
			fmt.Fprintf(os.Stderr, "hreysi: skill install warning: %v\n", serr)
		} else if len(written) > 0 {
			fmt.Println("  skill:    .claude/skills/expand — run it to narrate your day into the entry")
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
                    (--ambient wires a SessionEnd expansion hook; --ambient-stop adds Stop)
  hreysi capture    Append HEAD to today's entry (run by the hook; also manual)
  hreysi watch      Watch the reflog and capture every commit — any client, can't-miss
  hreysi doctor     Check that capture is actually wired and will fire
  hreysi version    Print version
  hreysi help       Show this help

The buildlog/ directory is the whole product: dated markdown, one timestamped
commit block each. Point a narrative-expansion skill, a content pipeline, or a
learning loop at it — hreysi only ever writes; consumers only ever read.
`)
}
