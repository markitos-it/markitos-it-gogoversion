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
	"testing"

	"github.com/go-git/go-git/v5/plumbing/object"
)

func TestParseCommit(t *testing.T) {
	cases := []struct {
		name     string
		message  string
		wantType string
		wantScope string
		wantSubject string
		wantBreaking bool
	}{
		{
			name:        "feat with scope",
			message:     "feat(auth): add oauth2 login",
			wantType:    "feat",
			wantScope:   "auth",
			wantSubject: "add oauth2 login",
		},
		{
			name:        "fix without scope",
			message:     "fix: correct null pointer",
			wantType:    "fix",
			wantScope:   "",
			wantSubject: "correct null pointer",
		},
		{
			name:         "breaking with exclamation",
			message:      "feat!: remove legacy API",
			wantType:     "feat",
			wantScope:    "",
			wantSubject:  "remove legacy API",
			wantBreaking: true,
		},
		{
			name:         "breaking with scope and exclamation",
			message:      "fix(api)!: breaking endpoint change",
			wantType:     "fix",
			wantScope:    "api",
			wantSubject:  "breaking endpoint change",
			wantBreaking: true,
		},
		{
			name:         "breaking in footer",
			message:      "feat: new thing\n\nBREAKING CHANGE: removed old endpoint",
			wantType:     "feat",
			wantScope:    "",
			wantSubject:  "new thing",
			wantBreaking: true,
		},
		{
			name:        "chore without scope",
			message:     "chore: update dependencies",
			wantType:    "chore",
			wantScope:   "",
			wantSubject: "update dependencies",
		},
		{
			name:        "non conventional commit",
			message:     "just a random commit message",
			wantType:    "",
			wantScope:   "",
			wantSubject: "just a random commit message",
		},
		{
			name:        "refactor with scope",
			message:     "refactor(core): simplify parser logic",
			wantType:    "refactor",
			wantScope:   "core",
			wantSubject: "simplify parser logic",
		},
		{
			name:        "message with extra whitespace",
			message:     "  fix: trim spaces  ",
			wantType:    "fix",
			wantScope:   "",
			wantSubject: "trim spaces",
		},
		{
			name:        "docs commit",
			message:     "docs(readme): update installation steps",
			wantType:    "docs",
			wantScope:   "readme",
			wantSubject: "update installation steps",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			raw := &object.Commit{Message: tc.message}
			got := parseCommit(raw)

			if got.Type != tc.wantType {
				t.Errorf("Type: got %q want %q", got.Type, tc.wantType)
			}
			if got.Scope != tc.wantScope {
				t.Errorf("Scope: got %q want %q", got.Scope, tc.wantScope)
			}
			if got.Subject != tc.wantSubject {
				t.Errorf("Subject: got %q want %q", got.Subject, tc.wantSubject)
			}
			if got.Breaking != tc.wantBreaking {
				t.Errorf("Breaking: got %v want %v", got.Breaking, tc.wantBreaking)
			}
		})
	}
}

func TestParseCommits(t *testing.T) {
	raw := []*object.Commit{
		{Message: "feat: one"},
		{Message: "fix: two"},
		{Message: "chore: three"},
	}

	got := parseCommits(raw)

	if len(got) != 3 {
		t.Fatalf("len: got %d want 3", len(got))
	}
	if got[0].Type != "feat" {
		t.Errorf("got[0].Type: got %q want %q", got[0].Type, "feat")
	}
	if got[1].Type != "fix" {
		t.Errorf("got[1].Type: got %q want %q", got[1].Type, "fix")
	}
	if got[2].Type != "chore" {
		t.Errorf("got[2].Type: got %q want %q", got[2].Type, "chore")
	}
}

func TestHasBreakingFooter(t *testing.T) {
	cases := []struct {
		name  string
		lines []string
		want  bool
	}{
		{"with breaking footer", []string{"feat: x", "BREAKING CHANGE: removed"}, true},
		{"without footer", []string{"feat: x"}, false},
		{"footer without breaking", []string{"feat: x", "just a note"}, false},
		{"empty footer", []string{"feat: x", ""}, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := hasBreakingFooter(tc.lines)
			if got != tc.want {
				t.Errorf("got %v want %v", got, tc.want)
			}
		})
	}
}

func TestExtractTypeAndScope(t *testing.T) {
	cases := []struct {
		name      string
		prefix    string
		wantType  string
		wantScope string
	}{
		{"with scope",    "feat(auth)",  "feat", "auth"},
		{"without scope", "fix",         "fix",  ""},
		{"unclosed paren","feat(auth",   "feat(auth", ""},
		{"empty scope",   "feat()",      "feat", ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotType, gotScope := extractTypeAndScope(tc.prefix)
			if gotType != tc.wantType {
				t.Errorf("type: got %q want %q", gotType, tc.wantType)
			}
			if gotScope != tc.wantScope {
				t.Errorf("scope: got %q want %q", gotScope, tc.wantScope)
			}
		})
	}
}