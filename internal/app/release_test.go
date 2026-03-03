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
	"strings"
	"testing"
)

func TestBumpVersion(t *testing.T) {
	cases := []struct {
		name       string
		current    string
		commits    []Commit
		wantNext   string
		wantReason string
	}{
		{
			name:       "breaking bumps major",
			current:    "1.2.3",
			commits:    []Commit{{Type: "feat", Breaking: true}},
			wantNext:   "2.0.0",
			wantReason: "breaking change detected → MAJOR bump",
		},
		{
			name:       "feat bumps minor",
			current:    "1.2.3",
			commits:    []Commit{{Type: "feat"}},
			wantNext:   "1.3.0",
			wantReason: "feature detected → MINOR bump",
		},
		{
			name:       "fix bumps patch",
			current:    "1.2.3",
			commits:    []Commit{{Type: "fix"}},
			wantNext:   "1.2.4",
			wantReason: "fix/chore detected → PATCH bump",
		},
		{
			name:       "chore bumps patch",
			current:    "1.2.3",
			commits:    []Commit{{Type: "chore"}},
			wantNext:   "1.2.4",
			wantReason: "fix/chore detected → PATCH bump",
		},
		{
			name:       "breaking wins over feat",
			current:    "1.2.3",
			commits:    []Commit{{Type: "feat"}, {Type: "fix", Breaking: true}},
			wantNext:   "2.0.0",
			wantReason: "breaking change detected → MAJOR bump",
		},
		{
			name:       "feat wins over fix",
			current:    "1.2.3",
			commits:    []Commit{{Type: "fix"}, {Type: "feat"}},
			wantNext:   "1.3.0",
			wantReason: "feature detected → MINOR bump",
		},
		{
			name:       "from zero version",
			current:    "0.0.0",
			commits:    []Commit{{Type: "feat"}},
			wantNext:   "0.1.0",
			wantReason: "feature detected → MINOR bump",
		},
		{
			name:       "invalid version falls back to 0.0.0",
			current:    "not-a-version",
			commits:    []Commit{{Type: "fix"}},
			wantNext:   "0.0.1",
			wantReason: "fix/chore detected → PATCH bump",
		},
		{
			name:       "empty commits bumps patch",
			current:    "1.0.0",
			commits:    []Commit{},
			wantNext:   "1.0.1",
			wantReason: "fix/chore detected → PATCH bump",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := buildReleaseResult(tc.current, tc.commits)

			wantNext := "v" + tc.wantNext
			if result.Next != wantNext {
				t.Errorf("Next: got %q want %q", result.Next, wantNext)
			}
			if result.Reason != tc.wantReason {
				t.Errorf("Reason: got %q want %q", result.Reason, tc.wantReason)
			}
			wantPrev := "v" + tc.current
			if tc.current == "not-a-version" {
				wantPrev = "v0.0.0"
			}
			if result.Previous != wantPrev {
				t.Errorf("Previous: got %q want %q", result.Previous, wantPrev)
			}
		})
	}
}

func TestAnyBreaking(t *testing.T) {
	cases := []struct {
		name    string
		commits []Commit
		want    bool
	}{
		{"one breaking", []Commit{{Breaking: true}}, true},
		{"none breaking", []Commit{{Breaking: false}, {Breaking: false}}, false},
		{"mixed", []Commit{{Breaking: false}, {Breaking: true}}, true},
		{"empty", []Commit{}, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := anyBreaking(tc.commits)
			if got != tc.want {
				t.Errorf("got %v want %v", got, tc.want)
			}
		})
	}
}

func TestAnyOfType(t *testing.T) {
	cases := []struct {
		name    string
		commits []Commit
		t       string
		want    bool
	}{
		{"found feat", []Commit{{Type: "feat"}, {Type: "fix"}}, "feat", true},
		{"not found", []Commit{{Type: "fix"}, {Type: "chore"}}, "feat", false},
		{"empty commits", []Commit{}, "feat", false},
		{"exact match", []Commit{{Type: "feature"}}, "feat", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := anyOfType(tc.commits, tc.t)
			if got != tc.want {
				t.Errorf("got %v want %v", got, tc.want)
			}
		})
	}
}

func TestFormatScope(t *testing.T) {
	cases := []struct {
		scope string
		want  string
	}{
		{"auth", "(auth) "},
		{"", ""},
		{"api", "(api) "},
	}

	for _, tc := range cases {
		t.Run(tc.scope, func(t *testing.T) {
			got := formatScope(tc.scope)
			if got != tc.want {
				t.Errorf("got %q want %q", got, tc.want)
			}
		})
	}
}

func TestFormatBreaking(t *testing.T) {
	if got := formatBreaking(true); got != " ⚠️ BREAKING" {
		t.Errorf("got %q want %q", got, " ⚠️ BREAKING")
	}
	if got := formatBreaking(false); got != "" {
		t.Errorf("got %q want %q", got, "")
	}
}

func TestFormatCommitLine(t *testing.T) {
	commit := Commit{Type: "feat", Scope: "api", Subject: "add endpoint", Breaking: true}
	got := formatCommitLine(commit)
	want := "feat: (api) add endpoint ⚠️ BREAKING"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestFormatCommitLineColored(t *testing.T) {
	paint := func(s, code string) string { return fmt.Sprintf("[%s]%s", code, s) }

	tests := []struct {
		name   string
		commit Commit
		code   string
	}{
		{name: "breaking wins", commit: Commit{Type: "fix", Subject: "x", Breaking: true}, code: ansiRed},
		{name: "feat", commit: Commit{Type: "feat", Subject: "x"}, code: ansiGreen},
		{name: "fix", commit: Commit{Type: "fix", Subject: "x"}, code: ansiYellow},
		{name: "perf", commit: Commit{Type: "perf", Subject: "x"}, code: ansiBlue},
		{name: "refactor", commit: Commit{Type: "refactor", Subject: "x"}, code: ansiMagenta},
		{name: "docs", commit: Commit{Type: "docs", Subject: "x"}, code: ansiBoldCyan},
		{name: "chore", commit: Commit{Type: "chore", Subject: "x"}, code: ansiBold},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := formatCommitLineColored(tc.commit, paint)
			expectedPrefix := fmt.Sprintf("[%s]", tc.code)
			if !strings.HasPrefix(got, expectedPrefix) {
				t.Fatalf("got %q want prefix %q", got, expectedPrefix)
			}
		})
	}

	t.Run("unknown no paint", func(t *testing.T) {
		commit := Commit{Type: "unknown", Subject: "x"}
		got := formatCommitLineColored(commit, paint)
		if strings.HasPrefix(got, "[") {
			t.Fatalf("expected no paint for unknown type, got %q", got)
		}
	})
}

func TestPrintSummaryOutput(t *testing.T) {
	origStdout := os.Stdout
	defer func() { os.Stdout = origStdout }()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stdout = w

	printSummary(ReleaseResult{
		Previous: "v0.1.0",
		Next:     "v0.1.1",
		Reason:   "fix/chore detected → PATCH bump",
		Commits: []Commit{
			{Hash: "abc1234", Type: "fix", Subject: "fix bug"},
		},
	})

	w.Close()
	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	for _, s := range []string{"Previous version", "Bump reason", "New version", "Included commits", "abc1234", "fix: fix bug"} {
		if !strings.Contains(output, s) {
			t.Fatalf("expected output to contain %q, got %q", s, output)
		}
	}
}
