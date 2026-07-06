// Command hreysi is an ambient buildlog capture tool: every git commit is
// appended to a dated markdown journal, with no ceremony and nothing to
// remember. The journal directory is the whole product — point anything at it.
package main

import (
	"embed"
	"fmt"
	"os"

	"github.com/Peleke/hreysi/internal/entry"
	"github.com/Peleke/hreysi/internal/gitx"
	"github.com/Peleke/hreysi/internal/scaffold"
	"github.com/Peleke/hreysi/internal/skillpack"
)

// version is overridden at release time via -ldflags "-X main.version=...".
var version = "dev"

// skillFS carries the skills that ship with hreysi (dropped into
// .claude/skills/ by `hreysi init`).
//
//go:embed skills
var skillFS embed.FS

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
	for _, a := range os.Args[2:] {
		if a == "--no-skill" {
			noSkill = true
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

	if !noSkill {
		if written, serr := skillpack.Install(res.Root, skillFS); serr != nil {
			fmt.Fprintf(os.Stderr, "hreysi: skill install warning: %v\n", serr)
		} else if len(written) > 0 {
			fmt.Println("  skill:    .claude/skills/expand — run it to narrate your day into the entry")
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
	info, err := gitx.Head(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "hreysi: could not read HEAD: %v\n", err)
		return 1
	}
	path, err := entry.Append(root, info)
	if err != nil {
		fmt.Fprintf(os.Stderr, "hreysi: could not write entry: %v\n", err)
		return 1
	}
	fmt.Printf("hreysi: captured %s → %s\n", info.Hash, path)
	return 0
}

func usage() {
	fmt.Print(`hreysi — ambient buildlog capture

Every git commit, appended to a dated journal. No ceremony.

USAGE:
  hreysi init       Scaffold buildlog/ and install the post-commit capture hook
  hreysi capture    Append HEAD to today's entry (run by the hook; also manual)
  hreysi version    Print version
  hreysi help       Show this help

The buildlog/ directory is the whole product: dated markdown, one timestamped
commit block each. Point a narrative-expansion skill, a content pipeline, or a
learning loop at it — hreysi only ever writes; consumers only ever read.
`)
}
