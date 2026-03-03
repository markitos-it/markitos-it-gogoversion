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
	"bufio"
	"flag"
	"fmt"
	"os"
	"slices"
	"strings"
)

func Run(version string) {
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
	result := buildReleaseResult(currentVersion, commits)

	if !cfg.DryRun && isInteractiveTerminal() {
		if !confirmDefaultRun(cfg, result) {
			fmt.Println("ℹ  Operación cancelada.")
			os.Exit(0)
		}
	}

	commitMessage := suggestedCommitMessage(result)
	if !cfg.DryRun && !cfg.NoChangelog && isInteractiveTerminal() {
		chosen, ok := askCommitMessage(result)
		if !ok {
			fmt.Println("ℹ  Operación cancelada.")
			os.Exit(0)
		}
		commitMessage = chosen
	}

	printSummary(result)

	if cfg.DryRun {
		fmt.Println("ℹ  --dry-run activo — sin cambios.")
		os.Exit(0)
	}

	if !cfg.NoChangelog {
		exitOnError(writeChangelog(cfg.RepoPath, result), "escribiendo CHANGELOG")
		fmt.Println("✔  CHANGELOG.md actualizado")
		exitOnError(addAndCommitChangelog(repo, commitMessage), "creando commit de release")
		fmt.Printf("✔  Commit creado: %s\n", commitMessage)
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

func confirmDefaultRun(cfg Config, result ReleaseResult) bool {
	suggestedCommit := suggestedCommitMessage(result)

	fmt.Println("╔══════════════════════════════════════════════════════╗")
	fmt.Println("║                  gogoversion · ggv                   ║")
	fmt.Println("╠══════════════════════════════════════════════════════╣")
	fmt.Println("║ Plan de release automático                            ║")
	fmt.Printf("║  • nuevo tag: %-40s║\n", result.Next)
	if !cfg.NoChangelog {
		fmt.Println("║  • actualizar CHANGELOG.md                           ║")
		fmt.Println("║  • commit sugerido:                                  ║")
		fmt.Printf("║    git add CHANGELOG.md                              ║\n")
		fmt.Printf("║    git commit -m %q\n", suggestedCommit)
	}
	if !cfg.NoTag {
		fmt.Println("║  • crear el nuevo tag git                            ║")
	}
	fmt.Println("║                                                      ║")
	fmt.Println("║ Plantillas commit (elige una):                       ║")
	fmt.Println("║  • feat: <mensaje>                                   ║")
	fmt.Println("║  • feat!: <mensaje>                                  ║")
	fmt.Println("║  • fix: <mensaje>                                    ║")
	fmt.Println("║  • fix!: <mensaje>                                   ║")
	fmt.Println("║  • perf: <mensaje>                                   ║")
	fmt.Println("║  • perf!: <mensaje>                                  ║")
	fmt.Println("║  • refactor: <mensaje>                               ║")
	fmt.Println("║  • refactor!: <mensaje>                              ║")
	fmt.Println("║  • docs: <mensaje>                                   ║")
	fmt.Println("║  • docs!: <mensaje>                                  ║")
	fmt.Println("║  • chore: <mensaje>                                  ║")
	fmt.Println("║  • chore!: <mensaje>                                 ║")
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

func suggestedCommitMessage(result ReleaseResult) string {
	if anyBreaking(result.Commits) {
		return fmt.Sprintf("feat(release)!: prepara release %s", result.Next)
	}
	if anyOfType(result.Commits, "feat") {
		return fmt.Sprintf("feat(release): prepara release %s", result.Next)
	}
	return fmt.Sprintf("fix(release): prepara release %s", result.Next)
}

func askCommitMessage(result ReleaseResult) (string, bool) {
	reader := bufio.NewReader(os.Stdin)
	defaultType := defaultReleaseCommitType(result)
	defaultBang := anyBreaking(result.Commits)
	defaultSubject := fmt.Sprintf("prepara release %s", result.Next)

	fmt.Println("\nConfigura el commit del release:")
	fmt.Printf("Tipo [%s]: ", defaultType)
	typeInput, err := reader.ReadString('\n')
	if err != nil {
		return "", false
	}
	selectedType := strings.TrimSpace(strings.ToLower(typeInput))
	if selectedType == "" {
		selectedType = defaultType
	}

	breaking := defaultBang
	if strings.HasSuffix(selectedType, "!") {
		breaking = true
		selectedType = strings.TrimSuffix(selectedType, "!")
	}

	if !isValidCommitType(selectedType) {
		fmt.Printf("✖  Tipo inválido: %q\n", selectedType)
		fmt.Println("   Tipos válidos: feat, fix, perf, refactor, docs, chore (opcional !)")
		return "", false
	}

	fmt.Printf("Mensaje [%s]: ", defaultSubject)
	subjectInput, err := reader.ReadString('\n')
	if err != nil {
		return "", false
	}
	subject := strings.TrimSpace(subjectInput)
	if subject == "" {
		subject = defaultSubject
	}

	if strings.EqualFold(subject, "cancel") || strings.EqualFold(subject, "cancelar") {
		return "", false
	}

	bang := ""
	if breaking {
		bang = "!"
	}

	return fmt.Sprintf("%s(release)%s: %s", selectedType, bang, subject), true
}

func defaultReleaseCommitType(result ReleaseResult) string {
	if anyOfType(result.Commits, "feat") {
		return "feat"
	}
	return "fix"
}

func isValidCommitType(commitType string) bool {
	allowed := []string{"feat", "fix", "perf", "refactor", "docs", "chore"}
	return slices.Contains(allowed, commitType)
}

func exitOnError(err error, context string) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "✖  Error %s: %v\n", context, err)
		os.Exit(1)
	}
}
