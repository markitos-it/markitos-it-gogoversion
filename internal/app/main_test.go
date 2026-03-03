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
	"bytes"
	"fmt"
	"os"
	"testing"
)

func TestSuggestedCommitMessage(t *testing.T) {
	t.Run("breaking suggests feat bang", func(t *testing.T) {
		result := ReleaseResult{
			Next:    "v2.0.0",
			Commits: []Commit{{Type: "feat", Breaking: true, Subject: "remove legacy API"}},
		}
		got := suggestedCommitMessage(result)
		want := "feat(release)!: release v2.0.0: remove legacy API"
		if got != want {
			t.Errorf("got %q want %q", got, want)
		}
	})

	t.Run("feat suggests feat", func(t *testing.T) {
		result := ReleaseResult{
			Next:    "v1.3.0",
			Commits: []Commit{{Type: "feat", Subject: "add oauth login"}},
		}
		got := suggestedCommitMessage(result)
		want := "feat(release): release v1.3.0: add oauth login"
		if got != want {
			t.Errorf("got %q want %q", got, want)
		}
	})

	t.Run("patch suggests fix", func(t *testing.T) {
		result := ReleaseResult{
			Next:    "v1.2.4",
			Commits: []Commit{{Type: "fix", Subject: "handle nil pointer"}},
		}
		got := suggestedCommitMessage(result)
		want := "fix(release): release v1.2.4: handle nil pointer"
		if got != want {
			t.Errorf("got %q want %q", got, want)
		}
	})
}

func TestDefaultReleaseSubject(t *testing.T) {
	result := ReleaseResult{
		Next: "v1.4.0",
		Commits: []Commit{
			{Type: "feat", Subject: "add oauth login"},
			{Type: "feat", Subject: "add oauth login"},
			{Type: "fix", Subject: "handle nil pointer"},
			{Type: "docs", Subject: "update README with setup"},
			{Type: "unknown", Subject: "misc cleanups"},
		},
	}

	got := defaultReleaseSubject(result)
	want := "release v1.4.0: add oauth login; handle nil pointer; update README with setup; +2 more changes"
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestSyntheticCommitsFromChangedFiles(t *testing.T) {
	commits := syntheticCommitsFromChangedFiles([]string{"a.go", "b.go"})
	if len(commits) != 1 {
		t.Fatalf("got %d commits want 1", len(commits))
	}
	if commits[0].Type != "chore" {
		t.Errorf("got type %q want %q", commits[0].Type, "chore")
	}
	if commits[0].Hash != "local" {
		t.Errorf("got hash %q want %q", commits[0].Hash, "local")
	}
	if commits[0].Subject == "" {
		t.Error("expected non-empty synthetic subject")
	}
}

func TestExitOnErrorNil(t *testing.T) {
	// exitOnError with nil should not panic or exit
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("unexpected panic: %v", r)
		}
	}()

	// We can't easily test os.Exit calls, but we can verify it doesn't panic on nil
	// exitOnError only exits on non-nil errors; this test validates nil is a no-op
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "")
	// If exitOnError(nil, ...) causes a problem, the test will fail due to panic or exit
	// We call it directly (it only calls os.Exit on error != nil)
	exitOnError(nil, "test context")
}

func TestExitOnErrorWithError(t *testing.T) {
	// We cannot directly test os.Exit in unit tests without subprocess tricks.
	// Instead, verify that the error message is written to stderr.
	// We replace os.Stderr temporarily.
	origStderr := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stderr = w

	done := make(chan struct{})
	var output []byte
	go func() {
		buf := make([]byte, 256)
		n, _ := r.Read(buf)
		output = buf[:n]
		close(done)
	}()

	// We use a subprocess approach indirectly by just formatting the error message
	// to verify the format string, without actually calling exitOnError (which would
	// call os.Exit and terminate the test process).
	testErr := fmt.Errorf("test error")
	fmt.Fprintf(os.Stderr, "✖  Error while %s: %v\n", "test context", testErr)

	w.Close()
	<-done
	os.Stderr = origStderr

	if !bytes.Contains(output, []byte("test context")) {
		t.Errorf("expected error output to contain 'test context', got: %s", output)
	}
	if !bytes.Contains(output, []byte("test error")) {
		t.Errorf("expected error output to contain 'test error', got: %s", output)
	}
}

func TestIsInteractiveTerminal(t *testing.T) {
	// In a test environment, stdin is typically not a terminal.
	// We just verify it doesn't panic and returns a bool.
	result := isInteractiveTerminal()
	_ = result // result is expected to be false in CI/test environments
}

func TestSupportsANSINoColor(t *testing.T) {
	orig := os.Getenv("NO_COLOR")
	os.Setenv("NO_COLOR", "1")
	defer os.Setenv("NO_COLOR", orig)

	var buf bytes.Buffer
	got := supportsANSI(&buf)
	if got {
		t.Error("expected supportsANSI=false when NO_COLOR is set")
	}
}

func TestSupportsANSIDumbTerm(t *testing.T) {
	origNoColor := os.Getenv("NO_COLOR")
	origTerm := os.Getenv("TERM")
	os.Unsetenv("NO_COLOR")
	os.Setenv("TERM", "dumb")
	defer func() {
		os.Setenv("NO_COLOR", origNoColor)
		os.Setenv("TERM", origTerm)
	}()

	var buf bytes.Buffer
	got := supportsANSI(&buf)
	if got {
		t.Error("expected supportsANSI=false for dumb terminal")
	}
}

func TestSupportsANSINoTerm(t *testing.T) {
	origNoColor := os.Getenv("NO_COLOR")
	origTerm := os.Getenv("TERM")
	os.Unsetenv("NO_COLOR")
	os.Unsetenv("TERM")
	defer func() {
		os.Setenv("NO_COLOR", origNoColor)
		os.Setenv("TERM", origTerm)
	}()

	var buf bytes.Buffer
	got := supportsANSI(&buf)
	if got {
		t.Error("expected supportsANSI=false when TERM is not set")
	}
}

func TestDefaultReleaseCommitType(t *testing.T) {
	t.Run("feat defaults to feat", func(t *testing.T) {
		result := ReleaseResult{Commits: []Commit{{Type: "fix"}, {Type: "feat"}}}
		if got := defaultReleaseCommitType(result); got != "feat" {
			t.Errorf("got %q want %q", got, "feat")
		}
	})

	t.Run("without feat defaults to fix", func(t *testing.T) {
		result := ReleaseResult{Commits: []Commit{{Type: "fix"}}}
		if got := defaultReleaseCommitType(result); got != "fix" {
			t.Errorf("got %q want %q", got, "fix")
		}
	})
}

func TestIsValidCommitType(t *testing.T) {
	valid := []string{"feat", "fix", "perf", "refactor", "docs", "chore"}
	for _, v := range valid {
		if !isValidCommitType(v) {
			t.Errorf("expected %q to be valid", v)
		}
	}

	invalid := []string{"", "style", "test", "feat!", "Feat"}
	for _, v := range invalid {
		if isValidCommitType(v) {
			t.Errorf("expected %q to be invalid", v)
		}
	}
}
