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
	"strings"

	"github.com/go-git/go-git/v5/plumbing/object"
)

type Commit struct {
	Hash     string
	Type     string
	Scope    string
	Subject  string
	Breaking bool
}

func parseCommits(raw []*object.Commit) []Commit {
	commits := make([]Commit, 0, len(raw))
	for _, c := range raw {
		commits = append(commits, parseCommit(c))
	}
	return commits
}

func parseCommit(c *object.Commit) Commit {
	lines  := splitMessage(c.Message)
	header := lines[0]

	commit := Commit{
		Hash:     c.Hash.String()[:7],
		Subject:  header,
		Breaking: hasBreakingFooter(lines),
	}

	colonIdx := strings.Index(header, ":")
	if colonIdx == -1 {
		return commit
	}

	prefix  := header[:colonIdx]
	commit.Subject = strings.TrimSpace(header[colonIdx+1:])

	if strings.HasSuffix(prefix, "!") {
		commit.Breaking = true
		prefix = strings.TrimSuffix(prefix, "!")
	}

	commit.Type, commit.Scope = extractTypeAndScope(prefix)
	return commit
}

func splitMessage(msg string) []string {
	return strings.SplitN(strings.TrimSpace(msg), "\n", 2)
}

func hasBreakingFooter(lines []string) bool {
	return len(lines) > 1 && strings.Contains(lines[1], "BREAKING CHANGE:")
}

func extractTypeAndScope(prefix string) (string, string) {
	openIdx := strings.Index(prefix, "(")
	if openIdx == -1 {
		return prefix, ""
	}
	closeIdx := strings.Index(prefix, ")")
	if closeIdx == -1 {
		return prefix, ""
	}
	return prefix[:openIdx], prefix[openIdx+1 : closeIdx]
}