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

	"github.com/Masterminds/semver/v3"
)

type ReleaseResult struct {
	Previous string
	Next     string
	Reason   string
	Commits  []Commit
}

func buildReleaseResult(current string, commits []Commit) ReleaseResult {
	v, err := semver.NewVersion(current)
	if err != nil {
		v, _ = semver.NewVersion("0.0.0")
	}

	result := ReleaseResult{
		Previous: "v" + v.String(),
		Commits:  commits,
	}

	next, reason := bumpVersion(v, commits)
	result.Next   = "v" + next
	result.Reason = reason
	return result
}

func bumpVersion(v *semver.Version, commits []Commit) (string, string) {
	major := v.Major()
	minor := v.Minor()
	patch := v.Patch()

	switch {
	case anyBreaking(commits):
		return fmt.Sprintf("%d.0.0", major+1), "BREAKING CHANGE detectado → bump MAJOR"
	case anyOfType(commits, "feat"):
		return fmt.Sprintf("%d.%d.0", major, minor+1), "feat detectado → bump MINOR"
	default:
		return fmt.Sprintf("%d.%d.%d", major, minor, patch+1), "fix/chore detectado → bump PATCH"
	}
}

func anyBreaking(commits []Commit) bool {
	for _, c := range commits {
		if c.Breaking {
			return true
		}
	}
	return false
}

func anyOfType(commits []Commit, t string) bool {
	for _, c := range commits {
		if c.Type == t {
			return true
		}
	}
	return false
}

func printSummary(result ReleaseResult) {
	fmt.Printf("▸  Versión anterior: %s\n", result.Previous)
	fmt.Printf("▸  Razón del bump:   %s\n", result.Reason)
	fmt.Printf("▸  Versión nueva:    %s\n\n", result.Next)
	fmt.Println("Commits incluidos:")
	for _, c := range result.Commits {
		fmt.Printf("  [%s] %s\n", c.Hash, formatCommitLine(c))
	}
	fmt.Println()
}

func formatCommitLine(c Commit) string {
	return fmt.Sprintf("%s: %s%s%s", c.Type, formatScope(c.Scope), c.Subject, formatBreaking(c.Breaking))
}

func formatScope(scope string) string {
	if scope == "" {
		return ""
	}
	return fmt.Sprintf("(%s) ", scope)
}

func formatBreaking(breaking bool) string {
	if !breaking {
		return ""
	}
	return " ⚠️ BREAKING"
}