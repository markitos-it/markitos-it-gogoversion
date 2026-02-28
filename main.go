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
	"fmt"
	"os"
)

func main() {
	cfg := newConfig()

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
		exitOnError(writeChangelog(result), "escribiendo CHANGELOG")
		fmt.Println("✔  CHANGELOG.md actualizado")
	}

	if !cfg.NoTag {
		exitOnError(createTag(repo, result.Next), "creando tag")
		fmt.Printf("✔  Tag %s creado\n", result.Next)
	}

	fmt.Printf("\n✅ Release %s lista\n", result.Next)
}

func exitOnError(err error, context string) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "✖  Error %s: %v\n", context, err)
		os.Exit(1)
	}
}