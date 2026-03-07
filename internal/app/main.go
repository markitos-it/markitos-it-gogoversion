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
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/go-git/go-git/v5"
	"github.com/manifoldco/promptui"
)

var (
	exitFunc           = os.Exit
	usageFunc          = func() { flag.Usage() }
	sleepFunc          = time.Sleep
	newConfigFunc      = newConfig
	askCommitMessageFn = askCommitMessage
	isInteractiveFn    = isInteractiveTerminal
	confirmReleaseFn   = confirmReleaseExecution
	selectCommitOption = func(selector *promptui.Select) (int, error) {
		idx, _, err := selector.Run()
		return idx, err
	}
	promptCommitMessage = func(prompt *promptui.Prompt) (string, error) {
		return prompt.Run()
	}
)

func Run(version string) {
	// Subcommand routing: checked before flag.Parse so that "install" and
	// "uninstall" are intercepted early. Note: global flags placed before the
	// subcommand (e.g. --verbose install) are NOT supported by this pattern.
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "install":
			installAliases()
			exitFunc(0)
			return
		case "uninstall":
			uninstallAliases()
			exitFunc(0)
			return
		}
	}

	cfg := newConfigFunc()

	if cfg.ShowHelp {
		usageFunc()
		exitFunc(0)
	}

	if cfg.ShowVersion {
		fmt.Println(version)
		exitFunc(0)
	}

	if cfg.Undo {
		exitOnError(undoLastRelease(cfg.RepoPath), "undoing last release")
		exitFunc(0)
	}

	repo, err := openRepository(cfg.RepoPath)
	exitOnError(err, "opening repository")

	currentVersion, err := latestTag(repo)
	exitOnError(err, "getting latest tag")

	contextInfo, err := collectReleaseContext(repo, cfg, version)
	exitOnError(err, "collecting release context")
	printReleaseContext(contextInfo)

	rawCommits, err := commitsSinceTag(repo, currentVersion)
	exitOnError(err, "reading commits")

	commits := parseCommits(rawCommits)
	if len(commits) == 0 {
		if len(contextInfo.ChangedFiles) > 0 {
			commits = syntheticCommitsFromChangedFiles(contextInfo.ChangedFiles)
			fmt.Printf("ℹ  No commits since latest tag; using %d local changed files for release planning.\n\n", len(contextInfo.ChangedFiles))
		} else {
			fmt.Println("⊘  No commits since latest tag and no local changes.")
			exitFunc(0)
		}
	}

	result := buildReleaseResult(currentVersion, commits)
	effectiveResult := resultForExecutionMode(result, currentVersion, cfg.NoTag)
	printSummary(effectiveResult)

	commitMessage := suggestedCommitMessage(effectiveResult)
	if !cfg.DryRun {
		chosen, ok := askCommitMessageFn(effectiveResult)
		if !ok {
			if isInteractiveFn() {
				fmt.Println("ℹ  Operation canceled.")
				exitFunc(0)
			}
			fmt.Println("ℹ  Interactive selector unavailable, using default commit message.")
		} else {
			commitMessage = chosen
		}
	}

	if adjusted, changed := applyCommitMessageBumpOverride(effectiveResult, currentVersion, cfg.NoTag, commitMessage); changed {
		effectiveResult = adjusted
		commitMessage = syncReleaseVersionInCommitMessage(commitMessage, effectiveResult.Next)
		fmt.Println("ℹ  Commit selection updated release bump.")
		fmt.Printf("▸  Bump reason:      %s\n", effectiveResult.Reason)
		fmt.Printf("▸  New version:      %s\n\n", effectiveResult.Next)
	}

	if cfg.DryRun {
		fmt.Println("ℹ  --dry-run active — no changes.")
		exitFunc(0)
	}

	if isInteractiveFn() {
		if !confirmReleaseFn(cfg, effectiveResult, commitMessage) {
			fmt.Println("ℹ  Operation canceled.")
			exitFunc(0)
		}
	}

	stagedCount := 0
	createdCommit := false
	createdTag := false

	runStep(1, 6, "git add")
	if !cfg.NoChangelog {
		exitOnError(writeChangelogForVersion(cfg.RepoPath, effectiveResult, effectiveResult.Next), "writing CHANGELOG")
		fmt.Println("✔  CHANGELOG.md updated")
	}
	stagedCount, err = addAllChangedFiles(repo)
	exitOnError(err, "staging files")
	fmt.Printf("✔  Staged %d changed files\n", stagedCount)

	runStep(2, 6, "git commit")
	if stagedCount > 0 {
		exitOnError(commitChanges(repo, commitMessage), "creating release commit")
		fmt.Printf("✔  Commit created: %s\n", commitMessage)
		createdCommit = true
	} else {
		fmt.Println("ℹ  No changed files to commit")
	}

	runStep(3, 6, "pull from origin")
	pulled, err := pullCurrentBranch(repo)
	exitOnError(err, "pulling latest changes from origin")
	if pulled {
		fmt.Println("✔  Pulled latest changes from origin")
	} else {
		fmt.Println("ℹ  Origin already up to date")
	}

	runStep(4, 6, "git tag")
	if !cfg.NoTag {
		exitOnError(createTag(repo, effectiveResult.Next), "creating tag")
		fmt.Printf("✔  Tag %s created\n", effectiveResult.Next)
		createdTag = true
	} else {
		fmt.Println("ℹ  Skipped tag: --no-tag enabled")
	}

	runStep(5, 6, "git push")
	if createdCommit || createdTag || pulled {
		exitOnError(pushRelease(repo, effectiveResult.Next, createdTag), "pushing changes to remote")
		fmt.Println("✔  Push to origin completed")
	} else {
		fmt.Println("ℹ  Nothing to push (no commit/tag created).")
	}

	runStep(6, 6, "git status")
	statusFiles, err := changedFiles(repo)
	exitOnError(err, "reading git status")
	if len(statusFiles) == 0 {
		fmt.Println("✔  Working tree is clean")
	} else {
		fmt.Printf("ℹ  Working tree has %d changed files:\n", len(statusFiles))
		for _, file := range statusFiles {
			fmt.Printf("    • %s\n", file)
		}
	}

	fmt.Printf("\n✅ Release %s ready\n", effectiveResult.Next)
}

func isInteractiveTerminal() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

func confirmReleaseExecution(cfg Config, result ReleaseResult, commitMessage string) bool {
	fmt.Println("Apply release actions")
	fmt.Printf("  - target version: %s\n", result.Next)
	if !cfg.NoChangelog {
		fmt.Println("  - write CHANGELOG.md: yes")
	} else {
		fmt.Println("  - write CHANGELOG.md: no")
	}
	fmt.Printf("  - create commit: yes (%s)\n", commitMessage)
	if !cfg.NoTag {
		fmt.Println("  - create tag: yes")
	} else {
		fmt.Println("  - create tag: no")
	}
	fmt.Println("  - pull from origin: yes")
	fmt.Println("  - push: yes")

	fmt.Print("Proceed? [y/N]: ")
	reader := bufio.NewReader(os.Stdin)
	answer, err := reader.ReadString('\n')
	if err != nil {
		return false
	}
	answer = strings.ToLower(strings.TrimSpace(answer))
	return answer == "y" || answer == "yes"
}

func suggestedCommitMessage(result ReleaseResult) string {
	commitType := defaultReleaseCommitType(result)
	bang := ""
	if anyBreaking(result.Commits) {
		bang = "!"
	}
	return fmt.Sprintf("%s(release)%s: %s", commitType, bang, defaultReleaseSubject(result))
}

func askCommitMessage(result ReleaseResult) (string, bool) {
	defaultType := defaultReleaseCommitType(result)
	defaultBang := anyBreaking(result.Commits)
	defaultSubject := defaultReleaseSubject(result)
	type commitOption struct {
		Value string
		Label string
	}

	fmt.Println("\nConfigure release commit:")
	choices := []commitOption{
		{Value: "feat", Label: "feat ✨ · new feature (MINOR)"},
		{Value: "feat!", Label: "feat! ⚠️ · BREAKING CHANGE (MAJOR)"},
		{Value: "fix", Label: "fix 🩹 · bug fix (PATCH)"},
		{Value: "fix!", Label: "fix! ⚠️ · breaking bug fix (MAJOR)"},
		{Value: "perf", Label: "perf 🚀 · performance improvement (PATCH)"},
		{Value: "perf!", Label: "perf! ⚠️ · breaking performance change (MAJOR)"},
		{Value: "refactor", Label: "refactor 🧱 · internal refactor (PATCH)"},
		{Value: "refactor!", Label: "refactor! ⚠️ · breaking refactor (MAJOR)"},
		{Value: "docs", Label: "docs 📝 · documentation (PATCH)"},
		{Value: "docs!", Label: "docs! ⚠️ · breaking docs change (MAJOR)"},
		{Value: "chore", Label: "chore 🔧 · maintenance (PATCH)"},
		{Value: "chore!", Label: "chore! ⚠️ · breaking maintenance change (MAJOR)"},
	}

	defaultChoice := defaultType
	if defaultBang {
		defaultChoice = defaultType + "!"
	}

	selector := promptui.Select{
		Label: "Commit type (↑/↓ and Enter)",
		Items: choices,
		Templates: &promptui.SelectTemplates{
			Label:    "{{ .Label | cyan }}",
			Active:   "▸ {{ .Label | green }}",
			Inactive: "  {{ .Label }}",
			Selected: "✔ {{ .Label | cyan }}",
		},
		Size: 8,
	}

	for i, item := range choices {
		if item.Value == defaultChoice {
			selector.CursorPos = i
			break
		}
	}

	idx, err := selectCommitOption(&selector)
	if err != nil {
		return "", false
	}

	selected := choices[idx].Value
	breaking := strings.HasSuffix(selected, "!")
	selectedType := strings.TrimSuffix(selected, "!")

	if !isValidCommitType(selectedType) {
		return "", false
	}

	previewResult := resultForSelectedCommitType(result, selectedType, breaking)
	defaultSubject = defaultReleaseSubject(previewResult)

	prompt := promptui.Prompt{
		Label:   "Message",
		Default: defaultSubject,
	}

	subject, err := promptCommitMessage(&prompt)
	if err != nil {
		return "", false
	}
	subject = strings.TrimSpace(subject)
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

func defaultReleaseSubject(result ReleaseResult) string {
	highlights := summarizeCommitSubjects(result.Commits, 3)
	if len(highlights) == 0 {
		return fmt.Sprintf("release %s", result.Next)
	}

	summary := strings.Join(highlights, "; ")
	if extra := countMeaningfulCommits(result.Commits) - len(highlights); extra > 0 {
		summary += fmt.Sprintf("; +%d more changes", extra)
	}

	return fmt.Sprintf("release %s: %s", result.Next, summary)
}

func summarizeCommitSubjects(commits []Commit, limit int) []string {
	if limit <= 0 {
		return nil
	}

	seen := map[string]struct{}{}
	highlights := make([]string, 0, limit)
	for _, commit := range commits {
		if shouldIgnoreForSummary(commit) {
			continue
		}
		subject := normalizeCommitSubject(commit.Subject)
		if subject == "" {
			continue
		}
		if _, ok := seen[subject]; ok {
			continue
		}
		seen[subject] = struct{}{}
		highlights = append(highlights, subject)
		if len(highlights) == limit {
			break
		}
	}

	return highlights
}

func shouldIgnoreForSummary(commit Commit) bool {
	if commit.Scope == "release" {
		return true
	}
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(commit.Subject)), "release v") {
		return true
	}
	return false
}

func normalizeCommitSubject(subject string) string {
	subject = strings.TrimSpace(subject)
	if subject == "" {
		return ""
	}

	if strings.HasPrefix(subject, "local working tree changes:") {
		changed := strings.TrimSpace(strings.TrimPrefix(subject, "local working tree changes:"))
		if changed == "" {
			return "local working tree changes"
		}

		parts := strings.Split(changed, ",")
		files := make([]string, 0, len(parts))
		for _, part := range parts {
			file := strings.TrimSpace(part)
			if file != "" {
				files = append(files, file)
			}
		}

		switch len(files) {
		case 0:
			return "local working tree changes"
		case 1:
			return fmt.Sprintf("update %s", files[0])
		case 2:
			return fmt.Sprintf("update %s and %s", files[0], files[1])
		default:
			return fmt.Sprintf("update %s, %s and %d more files", files[0], files[1], len(files)-2)
		}
	}

	const maxLen = 58
	if len(subject) > maxLen {
		return strings.TrimSpace(subject[:maxLen-1]) + "…"
	}

	return subject
}

func countMeaningfulCommits(commits []Commit) int {
	count := 0
	for _, commit := range commits {
		if normalizeCommitSubject(commit.Subject) != "" {
			count++
		}
	}
	return count
}

func syntheticCommitsFromChangedFiles(changedFiles []string) []Commit {
	subject := fmt.Sprintf("local working tree changes (%d files)", len(changedFiles))
	if len(changedFiles) > 0 {
		subject = fmt.Sprintf("local working tree changes: %s", strings.Join(changedFiles, ", "))
	}

	return []Commit{{
		Hash:    "local",
		Type:    "chore",
		Subject: subject,
	}}
}

type releaseContext struct {
	ToolVersion  string
	RepoPath     string
	Branch       string
	LatestTag    string
	ChangedFiles []string
	DryRun       bool
	NoTag        bool
	NoChangelog  bool
}

func collectReleaseContext(repo *git.Repository, cfg Config, version string) (releaseContext, error) {
	branch, err := currentBranchName(repo)
	if err != nil {
		return releaseContext{}, err
	}

	latest, err := latestTagName(repo)
	if err != nil {
		return releaseContext{}, err
	}
	if latest == "" {
		latest = "none"
	}

	changedFiles, err := changedFiles(repo)
	if err != nil {
		return releaseContext{}, err
	}

	return releaseContext{
		ToolVersion:  version,
		RepoPath:     cfg.RepoPath,
		Branch:       branch,
		LatestTag:    latest,
		ChangedFiles: changedFiles,
		DryRun:       cfg.DryRun,
		NoTag:        cfg.NoTag,
		NoChangelog:  cfg.NoChangelog,
	}, nil
}

func printReleaseContext(info releaseContext) {
	fmt.Println(styleLine("✨ markitos powered by gogoversion · ggv", ansiBoldCyan))
	fmt.Println("Release context")
	fmt.Printf("  - tool version: %s\n", info.ToolVersion)
	fmt.Printf("  - repo path: %s\n", info.RepoPath)
	fmt.Printf("  - branch: %s\n", info.Branch)
	fmt.Printf("  - latest tag: %s\n", info.LatestTag)
	if len(info.ChangedFiles) == 0 {
		fmt.Println("  - changed files: none")
	} else {
		fmt.Printf("  - changed files (%d):\n", len(info.ChangedFiles))
		for _, file := range info.ChangedFiles {
			fmt.Printf("    • %s\n", file)
		}
	}
	fmt.Printf("  - options: dry-run=%t no-tag=%t no-changelog=%t\n\n", info.DryRun, info.NoTag, info.NoChangelog)
}

func styleLine(text, code string) string {
	if !supportsANSI(os.Stdout) {
		return text
	}
	return code + text + ansiReset
}

func isValidCommitType(commitType string) bool {
	allowed := []string{"feat", "fix", "perf", "refactor", "docs", "chore"}
	return slices.Contains(allowed, commitType)
}

func exitOnError(err error, context string) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "✖  Error while %s: %v\n", context, err)
		exitFunc(1)
	}
}

func runStep(current, total int, name string) {
	fmt.Printf("\n▶ Step %d/%d: %s\n", current, total, name)
	sleepFunc(2 * time.Second)
}

func changelogBaseVersion(currentVersion string) string {
	v := strings.TrimSpace(currentVersion)
	if v == "" {
		v = "0.0.0"
	}
	if !strings.HasPrefix(v, "v") {
		v = "v" + v
	}
	return v
}

func resultForExecutionMode(result ReleaseResult, currentVersion string, noTag bool) ReleaseResult {
	if !noTag {
		return result
	}

	effective := result
	effective.Next = changelogBaseVersion(currentVersion)
	effective.Reason = "--no-tag enabled → keep current version (no new tag)"
	return effective
}

type bumpLevel int

const (
	bumpPatch bumpLevel = iota + 1
	bumpMinor
	bumpMajor
)

func applyCommitMessageBumpOverride(result ReleaseResult, currentVersion string, noTag bool, commitMessage string) (ReleaseResult, bool) {
	if noTag {
		return result, false
	}

	selectedLevel, ok := bumpLevelFromCommitMessage(commitMessage)
	if !ok {
		return result, false
	}

	computedLevel := bumpLevelFromCommits(result.Commits)
	if selectedLevel <= computedLevel {
		return result, false
	}

	base, err := semver.NewVersion(strings.TrimPrefix(result.Previous, "v"))
	if err != nil {
		base, err = semver.NewVersion(strings.TrimPrefix(changelogBaseVersion(currentVersion), "v"))
		if err != nil {
			base, _ = semver.NewVersion("0.0.0")
		}
	}

	next, reason := bumpVersionForLevel(base, selectedLevel)
	result.Next = "v" + next
	result.Reason = reason
	return result, true
}

func bumpLevelFromCommitMessage(commitMessage string) (bumpLevel, bool) {
	typeName, breaking, ok := parseReleaseCommitType(commitMessage)
	if !ok {
		return 0, false
	}

	if breaking {
		return bumpMajor, true
	}
	if typeName == "feat" {
		return bumpMinor, true
	}
	return bumpPatch, true
}

func bumpLevelFromCommits(commits []Commit) bumpLevel {
	if anyBreaking(commits) {
		return bumpMajor
	}
	if anyOfType(commits, "feat") {
		return bumpMinor
	}
	return bumpPatch
}

func bumpVersionForLevel(v *semver.Version, level bumpLevel) (string, string) {
	major := v.Major()
	minor := v.Minor()
	patch := v.Patch()

	switch level {
	case bumpMajor:
		return fmt.Sprintf("%d.0.0", major+1), "selected release commit marked BREAKING → MAJOR bump"
	case bumpMinor:
		return fmt.Sprintf("%d.%d.0", major, minor+1), "selected release commit type feat → MINOR bump"
	default:
		return fmt.Sprintf("%d.%d.%d", major, minor, patch+1), "selected release commit type fix/chore/perf/refactor/docs → PATCH bump"
	}
}

func parseReleaseCommitType(commitMessage string) (string, bool, bool) {
	message := strings.TrimSpace(commitMessage)
	if message == "" {
		return "", false, false
	}

	header := message
	if parts := strings.SplitN(message, ":", 2); len(parts) > 0 {
		header = strings.TrimSpace(parts[0])
	}

	breaking := strings.Contains(header, "!")
	left := header
	if idx := strings.Index(header, "("); idx >= 0 {
		left = strings.TrimSpace(header[:idx])
	}
	left = strings.TrimSuffix(left, "!")

	if !isValidCommitType(left) {
		return "", false, false
	}

	return left, breaking, true
}

func syncReleaseVersionInCommitMessage(commitMessage, targetVersion string) string {
	parts := strings.SplitN(strings.TrimSpace(commitMessage), ":", 2)
	if len(parts) != 2 {
		return commitMessage
	}

	header := strings.TrimSpace(parts[0])
	subject := strings.TrimSpace(parts[1])
	lower := strings.ToLower(subject)
	if !strings.HasPrefix(lower, "release v") {
		return commitMessage
	}

	rest := subject[len("release "):]
	end := 0
	for end < len(rest) && rest[end] != ' ' && rest[end] != ':' {
		end++
	}
	if end == 0 {
		return commitMessage
	}

	normalized := "release " + targetVersion + rest[end:]
	return fmt.Sprintf("%s: %s", header, normalized)
}

func resultForSelectedCommitType(result ReleaseResult, selectedType string, breaking bool) ReleaseResult {
	if strings.Contains(result.Reason, "--no-tag enabled") {
		return result
	}

	selectedLevel := bumpPatch
	if breaking {
		selectedLevel = bumpMajor
	} else if selectedType == "feat" {
		selectedLevel = bumpMinor
	}

	if selectedLevel <= bumpLevelFromCommits(result.Commits) {
		return result
	}

	base, err := semver.NewVersion(strings.TrimPrefix(result.Previous, "v"))
	if err != nil {
		return result
	}

	next, reason := bumpVersionForLevel(base, selectedLevel)
	updated := result
	updated.Next = "v" + next
	updated.Reason = reason
	return updated
}
