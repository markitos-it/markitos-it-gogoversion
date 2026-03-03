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
