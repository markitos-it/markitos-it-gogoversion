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
	"io"
	"os"
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
	return writeChangelogForVersion(repoPath, result, result.Next)
}

func writeChangelogForVersion(repoPath string, result ReleaseResult, version string) error {
	existing := readExistingChangelog(repoPath)
	root, err := os.OpenRoot(repoPath)
	if err != nil {
		return err
	}
	defer root.Close()
	f, err := root.OpenFile(changelogFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(buildEntry(version, result) + existing)
	return err
}

func readExistingChangelog(repoPath string) string {
	root, err := os.OpenRoot(repoPath)
	if err != nil {
		return ""
	}
	defer root.Close()
	f, err := root.Open(changelogFile)
	if err != nil {
		return ""
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		return ""
	}
	return string(data)
}

func removeChangelogEntry(repoPath, version string) (bool, error) {
	root, err := os.OpenRoot(repoPath)
	if err != nil {
		return false, err
	}
	defer root.Close()

	f, err := root.Open(changelogFile)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	data, err := io.ReadAll(f)
	f.Close()
	if err != nil {
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
	var updated string
	if nextRel == -1 {
		updated = strings.TrimSpace(content[:start])
		if updated != "" {
			updated += "\n"
		}
	} else {
		nextStart := start + len(header) + nextRel + 1
		updated = content[:start] + content[nextStart:]
		updated = strings.TrimLeft(updated, "\n")
	}

	wf, err := root.OpenFile(changelogFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return false, err
	}
	defer wf.Close()
	_, err = wf.WriteString(updated)
	return true, err
}

func buildEntry(version string, result ReleaseResult) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## %s (%s)\n\n", version, time.Now().Format("2006-01-02")))
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
