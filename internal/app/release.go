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
	"slices"

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
	result.Next = "v" + next
	result.Reason = reason
	return result
}

func bumpVersion(v *semver.Version, commits []Commit) (string, string) {
	major := v.Major()
	minor := v.Minor()
	patch := v.Patch()

	switch {
	case anyBreaking(commits):
		return fmt.Sprintf("%d.0.0", major+1), "breaking change detected → MAJOR bump"
	case anyOfType(commits, "feat"):
		return fmt.Sprintf("%d.%d.0", major, minor+1), "feature detected → MINOR bump"
	default:
		return fmt.Sprintf("%d.%d.%d", major, minor, patch+1), "fix/chore detected → PATCH bump"
	}
}

func anyBreaking(commits []Commit) bool {
	return slices.ContainsFunc(commits, func(c Commit) bool { return c.Breaking })
}

func anyOfType(commits []Commit, t string) bool {
	return slices.ContainsFunc(commits, func(c Commit) bool { return c.Type == t })
}

func printSummary(result ReleaseResult) {
	paint := colorizer(os.Stdout)
	fmt.Printf("▸  Previous version: %s\n", result.Previous)
	fmt.Printf("▸  Bump reason:      %s\n", result.Reason)
	fmt.Printf("▸  New version:      %s\n\n", result.Next)
	fmt.Println("Included commits:")
	for _, c := range result.Commits {
		fmt.Printf("  [%s] %s\n", c.Hash, formatCommitLineColored(c, paint))
	}
	fmt.Println()
}

func formatCommitLineColored(c Commit, paint func(string, string) string) string {
	line := formatCommitLine(c)
	if c.Breaking {
		return paint(line, ansiRed)
	}

	switch c.Type {
	case "feat":
		return paint(line, ansiGreen)
	case "fix":
		return paint(line, ansiYellow)
	case "perf":
		return paint(line, ansiBlue)
	case "refactor":
		return paint(line, ansiMagenta)
	case "docs":
		return paint(line, ansiBoldCyan)
	case "chore":
		return paint(line, ansiBold)
	default:
		return line
	}
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
