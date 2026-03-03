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
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const changelogFile = "CHANGELOG.md"

var groupOrder = []string{"feat", "fix", "perf", "refactor", "docs", "chore", "other"}

var groupLabels = map[string]string{
	"feat":     "✨ Features",
	"fix":      "🐛 Bug Fixes",
	"perf":     "⚡ Performance",
	"refactor": "♻️  Refactor",
	"docs":     "📚 Docs",
	"chore":    "🔧 Chores",
	"other":    "📦 Other",
}

func writeChangelog(repoPath string, result ReleaseResult) error {
	existing := readExistingChangelog(repoPath)
	return os.WriteFile(changelogPath(repoPath), []byte(buildEntry(result)+existing), 0644)
}

func readExistingChangelog(repoPath string) string {
	data, err := os.ReadFile(changelogPath(repoPath))
	if err != nil {
		return ""
	}
	return string(data)
}

func changelogPath(repoPath string) string {
	return filepath.Join(repoPath, changelogFile)
}

func removeChangelogEntry(repoPath, version string) (bool, error) {
	path := changelogPath(repoPath)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	content := string(data)
	header := "## " + version + " ("
	start := strings.Index(content, header)
	if start == -1 {
		return false, nil
	}

	rest := content[start:]
	nextRel := strings.Index(rest[len(header):], "\n## ")
	if nextRel == -1 {
		updated := strings.TrimSpace(content[:start])
		if updated != "" {
			updated += "\n"
		}
		return true, os.WriteFile(path, []byte(updated), 0644)
	}

	nextStart := start + len(header) + nextRel + 1
	updated := content[:start] + content[nextStart:]
	updated = strings.TrimLeft(updated, "\n")
	return true, os.WriteFile(path, []byte(updated), 0644)
}

func buildEntry(result ReleaseResult) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## %s (%s)\n\n", result.Next, time.Now().Format("2006-01-02")))
	for _, t := range groupOrder {
		writeGroup(&sb, t, filterByType(result.Commits, t))
	}
	return sb.String()
}

func filterByType(commits []Commit, t string) []Commit {
	var filtered []Commit
	for _, c := range commits {
		key := c.Type
		if _, ok := groupLabels[key]; !ok {
			key = "other"
		}
		if key == t {
			filtered = append(filtered, c)
		}
	}
	return filtered
}

func writeGroup(sb *strings.Builder, t string, commits []Commit) {
	if len(commits) == 0 {
		return
	}
	sb.WriteString(fmt.Sprintf("### %s\n\n", groupLabels[t]))
	for _, c := range commits {
		sb.WriteString(fmt.Sprintf("- %s%s ([`%s`])%s\n",
			formatScope(c.Scope), c.Subject, c.Hash, formatBreaking(c.Breaking)))
	}
	sb.WriteString("\n")
}
