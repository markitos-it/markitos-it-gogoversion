//#:[.'.]:>-==================================================================================
//#:[.'.]:>- Marco Antonio - markitos devsecops kulture
//#:[.'.]:>- The Way of the Artisan
//#:[.'.]:>- markitos.es.info@gmail.com
//#:[.'.]:>- 🌍 https://github.com/orgs/markitos-it/repositories
//#:[.'.]:>- 🌍 https://github.com/orgs/markitos-public/repositories
//#:[.'.]:>- 📺 https://www.youtube.com/@markitos_devsecops
//#:[.'.]:>- =================================================================================

package main

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

func newConfig() Config {
	cfg := Config{}
	flag.Usage = func() {
		out := flag.CommandLine.Output()
		paint := colorizer(out)
		name := filepath.Base(os.Args[0])
		fmt.Fprintf(out, "%s: %s [flags] <repo_path>\n\n", paint("Uso", ansiBoldCyan), paint(name, ansiBold))
		fmt.Fprintln(out, paint("Notas", ansiBoldCyan)+":")
		fmt.Fprintln(out, "  - repo_path es obligatorio y va siempre al final")
		fmt.Fprintln(out, "  - usa . para el repositorio actual")
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, paint("Ejemplos", ansiBoldCyan)+":")
		fmt.Fprintf(out, "  %s .\n", paint(name, ansiYellow))
		fmt.Fprintf(out, "  %s --dry-run .\n", paint(name, ansiYellow))
		fmt.Fprintf(out, "  %s --no-tag --no-changelog /ruta/a/repo\n", paint(name, ansiYellow))
		fmt.Fprintf(out, "  %s --undo .\n", paint(name, ansiYellow))
		fmt.Fprintf(out, "  %s --undo ../otro-repo\n\n", paint(name, ansiYellow))
		fmt.Fprintln(out, paint("Flags", ansiBoldCyan)+":")
		flag.PrintDefaults()
	}
	flag.BoolVar(&cfg.DryRun,        "dry-run",       false, "Muestra la versión sin escribir nada")
	flag.BoolVar(&cfg.NoTag,         "no-tag",        false, "No crea el tag git")
	flag.BoolVar(&cfg.NoChangelog,   "no-changelog",  false, "No escribe CHANGELOG.md")
	flag.BoolVar(&cfg.Undo,          "undo",          false, "Deshace el último release del repo indicado (elimina tag y entrada de CHANGELOG.md)")
	flag.BoolVar(&cfg.ShowHelp,      "h",             false, "Muestra esta ayuda")
	flag.BoolVar(&cfg.ShowHelp,      "help",          false, "Muestra esta ayuda")
	flag.BoolVar(&cfg.ShowVersion,   "version",       false, "Muestra la versión del binario")

	rawArgs := os.Args[1:]
	if len(rawArgs) == 1 && rawArgs[0] == "--" {
		cfg.ShowHelp = true
		return cfg
	}

	if err := flag.CommandLine.Parse(rawArgs); err != nil {
		os.Exit(2)
	}

	if cfg.ShowHelp || cfg.ShowVersion {
		return cfg
	}

	args := flag.Args()
	if len(args) != 1 {
		if cfg.Undo && len(args) == 0 {
			fmt.Fprintln(os.Stderr, "✖  Error: --undo requiere repo_path al final. Ejemplo: gogoversion --undo .")
			flag.Usage()
			os.Exit(2)
		}
		fmt.Fprintln(os.Stderr, "✖  Error: debes indicar exactamente un repo_path al final (usa . para el repo actual)")
		flag.Usage()
		os.Exit(2)
	}

	repoPath := args[0]
	if repoPath == "-" || repoPath == "--" {
		fmt.Fprintln(os.Stderr, "✖  Error: repo_path inválido")
		flag.Usage()
		os.Exit(2)
	}

	if rawArgs[len(rawArgs)-1] != repoPath {
		fmt.Fprintln(os.Stderr, "✖  Error: repo_path debe ir al final, después de todas las opciones")
		flag.Usage()
		os.Exit(2)
	}

	cfg.RepoPath = repoPath

	return cfg
}

const (
	ansiReset    = "\033[0m"
	ansiBold     = "\033[1m"
	ansiYellow   = "\033[33m"
	ansiBoldCyan = "\033[1;36m"
)

func colorizer(w io.Writer) func(string, string) string {
	if !supportsANSI(w) {
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