//#:[.'.]:>-==================================================================================
//#:[.'.]:>- Marco Antonio - markitos devsecops kulture
//#:[.'.]:>- The Way of the Artisan
//#:[.'.]:>- markitos.es.info@gmail.com
//#:[.'.]:>- 🌍 https://github.com/orgs/markitos-it/repositories
//#:[.'.]:>- 🌍 https://github.com/orgs/markitos-public/repositories
//#:[.'.]:>- 📺 https://www.youtube.com/@markitos_devsecops
//#:[.'.]:>- =================================================================================

package app

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Config struct {
	RepoPath    string
	DryRun      bool
	NoTag       bool
	NoChangelog bool
	Undo        bool
	ShowHelp    bool
	ShowVersion bool
}

var configExitFunc = os.Exit

func newConfig() Config {
	cfg := Config{}
	flag.Usage = func() {
		out := flag.CommandLine.Output()
		paint := colorizer(out)
		name := filepath.Base(os.Args[0])
		fmt.Fprintf(out, "%s: %s [flags] <repo_path>\n\n", paint("Usage", ansiBoldCyan), paint(name, ansiBold))
		fmt.Fprintln(out, paint("Notes", ansiBoldCyan)+":")
		fmt.Fprintln(out, "  - repo_path is required and must be the last argument")
		fmt.Fprintln(out, "  - use . for the current repository")
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, paint("Subcommands", ansiBoldCyan)+":")
		fmt.Fprintf(out, "  %s install    %s\n", paint(name, ansiYellow), paint("Create aliases: ggv, gogov (and gogoversion if needed)", ansiBlue))
		fmt.Fprintf(out, "  %s uninstall  %s\n\n", paint(name, ansiYellow), paint("Remove aliases: ggv, gogov (and gogoversion if needed)", ansiBlue))
		fmt.Fprintln(out, paint("Examples", ansiBoldCyan)+":")
		fmt.Fprintf(out, "  %s .\n", paint(name, ansiYellow))
		fmt.Fprintf(out, "  %s --dry-run .\n", paint(name, ansiYellow))
		fmt.Fprintf(out, "  %s --no-tag --no-changelog /path/to/repo\n", paint(name, ansiYellow))
		fmt.Fprintf(out, "  %s --undo .\n", paint(name, ansiYellow))
		fmt.Fprintf(out, "  %s --undo ../otro-repo\n", paint(name, ansiYellow))
		fmt.Fprintf(out, "  %s install\n\n", paint(name, ansiYellow))
		fmt.Fprintln(out, paint("Flags", ansiBoldCyan)+":")
		flag.PrintDefaults()
	}
	flag.BoolVar(&cfg.DryRun, "dry-run", false, "Show the computed version without writing anything")
	flag.BoolVar(&cfg.NoTag, "no-tag", false, "Do not create the git tag")
	flag.BoolVar(&cfg.NoChangelog, "no-changelog", false, "Do not write CHANGELOG.md")
	flag.BoolVar(&cfg.Undo, "undo", false, "Undo the last release of the given repo (remove tag and CHANGELOG.md entry)")
	flag.BoolVar(&cfg.ShowHelp, "h", false, "Show this help")
	flag.BoolVar(&cfg.ShowHelp, "help", false, "Show this help")
	flag.BoolVar(&cfg.ShowVersion, "version", false, "Show binary version")

	rawArgs := os.Args[1:]
	if len(rawArgs) == 0 {
		cfg.ShowHelp = true
		return cfg
	}

	if len(rawArgs) == 1 && rawArgs[0] == "--" {
		cfg.ShowHelp = true
		return cfg
	}

	if err := flag.CommandLine.Parse(rawArgs); err != nil {
		configExitFunc(2)
	}

	if cfg.ShowHelp || cfg.ShowVersion {
		return cfg
	}

	args := flag.Args()
	if len(args) != 1 {
		if cfg.Undo && len(args) == 0 {
			fmt.Fprintln(os.Stderr, "✖  Error: --undo requires repo_path at the end. Example: gogoversion --undo .")
			flag.Usage()
			configExitFunc(2)
		}
		fmt.Fprintln(os.Stderr, "✖  Error: provide exactly one repo_path at the end (use . for current repo)")
		flag.Usage()
		configExitFunc(2)
	}

	repoPath := args[0]
	if repoPath == "-" || repoPath == "--" {
		fmt.Fprintln(os.Stderr, "✖  Error: invalid repo_path")
		flag.Usage()
		configExitFunc(2)
	}

	if rawArgs[len(rawArgs)-1] != repoPath {
		fmt.Fprintln(os.Stderr, "✖  Error: repo_path must be last, after all options")
		flag.Usage()
		configExitFunc(2)
	}

	cfg.RepoPath = repoPath

	return cfg
}

const (
	ansiReset    = "\033[0m"
	ansiBold     = "\033[1m"
	ansiYellow   = "\033[33m"
	ansiGreen    = "\033[32m"
	ansiBlue     = "\033[34m"
	ansiMagenta  = "\033[35m"
	ansiRed      = "\033[31m"
	ansiBoldCyan = "\033[1;36m"
)

var supportsANSICheck = supportsANSI

func colorizer(w io.Writer) func(string, string) string {
	if !supportsANSICheck(w) {
		return func(s, _ string) string { return s }
	}
	return func(s, code string) string {
		return code + s + ansiReset
	}
}

func supportsANSI(w io.Writer) bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	term := os.Getenv("TERM")
	if term == "" || term == "dumb" {
		return false
	}
	file, ok := w.(*os.File)
	if !ok {
		return false
	}
	info, err := file.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}
