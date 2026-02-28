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
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
)

var version = "dev"

func main() {
	cfg := newConfig()

	if cfg.ShowHelp {
		flag.Usage()
		os.Exit(0)
	}

	if cfg.ShowVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	if cfg.Undo {
		exitOnError(undoLastRelease(cfg.RepoPath), "deshaciendo último release")
		os.Exit(0)
	}

	if !cfg.DryRun && isInteractiveTerminal() {
		if !confirmDefaultRun(cfg) {
			fmt.Println("ℹ  Operación cancelada.")
			os.Exit(0)
		}
	}

	repo, err := openRepository(cfg.RepoPath)
	exitOnError(err, "abriendo repositorio")

	currentVersion, err := latestTag(repo)
	exitOnError(err, "obteniendo último tag")

	rawCommits, err := commitsSinceTag(repo, currentVersion)
	exitOnError(err, "leyendo commits")

	if len(rawCommits) == 0 {
		fmt.Println("⊘  Sin commits desde el último tag.")
		os.Exit(0)
	}

	commits := parseCommits(rawCommits)
	result  := buildReleaseResult(currentVersion, commits)

	printSummary(result)

	if cfg.DryRun {
		fmt.Println("ℹ  --dry-run activo — sin cambios.")
		os.Exit(0)
	}

	if !cfg.NoChangelog {
		exitOnError(writeChangelog(cfg.RepoPath, result), "escribiendo CHANGELOG")
		fmt.Println("✔  CHANGELOG.md actualizado")
	}

	if !cfg.NoTag {
		exitOnError(createTag(repo, result.Next), "creando tag")
		fmt.Printf("✔  Tag %s creado\n", result.Next)
	}

	fmt.Printf("\n✅ Release %s lista\n", result.Next)
}

func isInteractiveTerminal() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

func confirmDefaultRun(cfg Config) bool {
	fmt.Println("╔══════════════════════════════════════════════════════╗")
	fmt.Println("║                  gogoversion · ggv                   ║")
	fmt.Println("╠══════════════════════════════════════════════════════╣")
	fmt.Println("║ Se ejecutará una release automática en este repo:    ║")
	fmt.Println("║  • leer commits desde el último tag                  ║")
	fmt.Println("║  • calcular la próxima versión (SemVer)              ║")
	if !cfg.NoChangelog {
		fmt.Println("║  • actualizar CHANGELOG.md                           ║")
	}
	if !cfg.NoTag {
		fmt.Println("║  • crear el nuevo tag git                            ║")
	}
	fmt.Println("╚══════════════════════════════════════════════════════╝")

	fmt.Print("¿Continuar? [y/N]: ")
	reader := bufio.NewReader(os.Stdin)
	answer, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	answer = strings.ToLower(strings.TrimSpace(answer))
	return answer == "y" || answer == "yes" || answer == "s" || answer == "si"
}

func exitOnError(err error, context string) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "✖  Error %s: %v\n", context, err)
		os.Exit(1)
	}
}
