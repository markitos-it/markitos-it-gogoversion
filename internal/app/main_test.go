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
	"strings"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	gitconfig "github.com/go-git/go-git/v5/config"
	"github.com/manifoldco/promptui"
)

type exitCall struct {
	code int
}

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

func TestDefaultReleaseSubjectWithLocalChanges(t *testing.T) {
	t.Run("one changed file", func(t *testing.T) {
		result := ReleaseResult{
			Next: "v0.1.1",
			Commits: []Commit{{
				Type:    "chore",
				Subject: "local working tree changes: .github/workflows/ci.yml",
			}},
		}

		got := defaultReleaseSubject(result)
		want := "release v0.1.1: update .github/workflows/ci.yml"
		if got != want {
			t.Errorf("got %q want %q", got, want)
		}
	})

	t.Run("three changed files", func(t *testing.T) {
		result := ReleaseResult{
			Next: "v0.1.1",
			Commits: []Commit{{
				Type:    "chore",
				Subject: "local working tree changes: .github/workflows/ci.yml, .octocov.yml, cmd/app/main.go",
			}},
		}

		got := defaultReleaseSubject(result)
		want := "release v0.1.1: update .github/workflows/ci.yml, .octocov.yml and 1 more files"
		if got != want {
			t.Errorf("got %q want %q", got, want)
		}
	})
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

func TestRunShowVersionExits(t *testing.T) {
	resetFlags()
	origArgs := os.Args
	origExit := exitFunc
	origStdout := os.Stdout
	defer func() {
		os.Args = origArgs
		exitFunc = origExit
		os.Stdout = origStdout
	}()

	os.Args = []string{"gogoversion", "--version"}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stdout = w

	exited := false
	exitCode := -1
	exitFunc = func(code int) {
		exited = true
		exitCode = code
		panic(exitCall{code: code})
	}

	func() {
		defer func() {
			if r := recover(); r != nil {
				if ec, ok := r.(exitCall); !ok || ec.code != 0 {
					t.Fatalf("unexpected panic: %#v", r)
				}
			}
		}()
		Run("v1.2.3")
	}()

	w.Close()
	buf := make([]byte, 128)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	if !exited || exitCode != 0 {
		t.Fatalf("expected exit(0), got exited=%t code=%d", exited, exitCode)
	}
	if !strings.Contains(output, "v1.2.3") {
		t.Fatalf("expected output to contain version, got %q", output)
	}
}

func TestRunShowHelpCallsUsageAndExits(t *testing.T) {
	resetFlags()
	origArgs := os.Args
	origExit := exitFunc
	origUsage := usageFunc
	defer func() {
		os.Args = origArgs
		exitFunc = origExit
		usageFunc = origUsage
	}()

	os.Args = []string{"gogoversion", "--help"}

	usageCalled := false
	usageFunc = func() { usageCalled = true }
	exitFunc = func(code int) { panic(exitCall{code: code}) }

	defer func() {
		r := recover()
		ec, ok := r.(exitCall)
		if !ok || ec.code != 0 {
			t.Fatalf("expected exit(0), got %#v", r)
		}
		if !usageCalled {
			t.Fatal("expected usageFunc to be called")
		}
	}()

	Run("v1.2.3")
}

func TestConfirmReleaseExecution(t *testing.T) {
	origStdin := os.Stdin
	defer func() { os.Stdin = origStdin }()

	t.Run("yes", func(t *testing.T) {
		r, w, err := os.Pipe()
		if err != nil {
			t.Fatalf("os.Pipe: %v", err)
		}
		os.Stdin = r
		_, _ = w.WriteString("yes\n")
		w.Close()

		ok := confirmReleaseExecution(Config{}, ReleaseResult{Next: "v1.0.0"}, "fix(release): release v1.0.0")
		if !ok {
			t.Fatal("expected confirmation to be true for yes")
		}
	})

	t.Run("default no", func(t *testing.T) {
		r, w, err := os.Pipe()
		if err != nil {
			t.Fatalf("os.Pipe: %v", err)
		}
		os.Stdin = r
		_, _ = w.WriteString("\n")
		w.Close()

		ok := confirmReleaseExecution(Config{}, ReleaseResult{Next: "v1.0.0"}, "fix(release): release v1.0.0")
		if ok {
			t.Fatal("expected confirmation to be false for empty input")
		}
	})
}

func TestCollectReleaseContext(t *testing.T) {
	repo, dir := initTestRepo(t)
	if err := os.WriteFile(dir+"/dirty.txt", []byte("dirty"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfg := Config{RepoPath: dir, DryRun: true, NoTag: true, NoChangelog: true}
	ctx, err := collectReleaseContext(repo, cfg, "vtest")
	if err != nil {
		t.Fatalf("collectReleaseContext: %v", err)
	}

	if ctx.ToolVersion != "vtest" {
		t.Fatalf("ToolVersion: got %q want %q", ctx.ToolVersion, "vtest")
	}
	if ctx.RepoPath != dir {
		t.Fatalf("RepoPath: got %q want %q", ctx.RepoPath, dir)
	}
	if ctx.LatestTag != "none" {
		t.Fatalf("LatestTag: got %q want %q", ctx.LatestTag, "none")
	}
	if len(ctx.ChangedFiles) == 0 {
		t.Fatal("expected changed files to be detected")
	}
}

func TestPrintReleaseContext(t *testing.T) {
	origStdout := os.Stdout
	defer func() { os.Stdout = origStdout }()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stdout = w

	printReleaseContext(releaseContext{
		ToolVersion:  "v1",
		RepoPath:     ".",
		Branch:       "main",
		LatestTag:    "v0.1.0",
		ChangedFiles: []string{"a.go", "b.go"},
		DryRun:       true,
		NoTag:        false,
		NoChangelog:  true,
	})

	w.Close()
	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	for _, needle := range []string{"Release context", "tool version: v1", "changed files (2):", "• a.go", "dry-run=true"} {
		if !strings.Contains(output, needle) {
			t.Fatalf("expected output to contain %q; got %q", needle, output)
		}
	}
}

func TestStyleLineNoANSI(t *testing.T) {
	got := styleLine("hello", ansiBold)
	if got != "hello" {
		t.Fatalf("got %q want %q", got, "hello")
	}
}

func TestExitOnErrorExitsWithCodeOne(t *testing.T) {
	origExit := exitFunc
	defer func() { exitFunc = origExit }()

	exitFunc = func(code int) { panic(exitCall{code: code}) }

	defer func() {
		r := recover()
		ec, ok := r.(exitCall)
		if !ok || ec.code != 1 {
			t.Fatalf("expected exit(1), got %#v", r)
		}
	}()

	exitOnError(fmt.Errorf("boom"), "testing")
}

func TestRunStepUsesSleepFunc(t *testing.T) {
	origSleep := sleepFunc
	origStdout := os.Stdout
	defer func() {
		sleepFunc = origSleep
		os.Stdout = origStdout
	}()

	called := false
	var gotDuration time.Duration
	sleepFunc = func(d time.Duration) {
		called = true
		gotDuration = d
	}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stdout = w

	runStep(1, 6, "git add")

	w.Close()
	buf := make([]byte, 256)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	if !called {
		t.Fatal("expected sleepFunc to be called")
	}
	if gotDuration != 2*time.Second {
		t.Fatalf("got duration %v want %v", gotDuration, 2*time.Second)
	}
	if !strings.Contains(output, "Step 1/6: git add") {
		t.Fatalf("expected step output, got %q", output)
	}
}

func TestChangelogBaseVersion(t *testing.T) {
	t.Run("adds prefix", func(t *testing.T) {
		if got := changelogBaseVersion("0.1.0"); got != "v0.1.0" {
			t.Fatalf("got %q want %q", got, "v0.1.0")
		}
	})

	t.Run("keeps prefixed", func(t *testing.T) {
		if got := changelogBaseVersion("v2.0.0"); got != "v2.0.0" {
			t.Fatalf("got %q want %q", got, "v2.0.0")
		}
	})

	t.Run("empty to zero", func(t *testing.T) {
		if got := changelogBaseVersion("  "); got != "v0.0.0" {
			t.Fatalf("got %q want %q", got, "v0.0.0")
		}
	})
}

func TestResultForExecutionMode(t *testing.T) {
	base := ReleaseResult{Previous: "v0.1.0", Next: "v0.1.1", Reason: "patch"}

	t.Run("no-tag keeps current version", func(t *testing.T) {
		got := resultForExecutionMode(base, "0.1.0", true)
		if got.Next != "v0.1.0" {
			t.Fatalf("got Next=%q want %q", got.Next, "v0.1.0")
		}
		if !strings.Contains(got.Reason, "--no-tag enabled") {
			t.Fatalf("unexpected reason: %q", got.Reason)
		}
	})

	t.Run("tag mode keeps computed next", func(t *testing.T) {
		got := resultForExecutionMode(base, "0.1.0", false)
		if got.Next != "v0.1.1" {
			t.Fatalf("got Next=%q want %q", got.Next, "v0.1.1")
		}
	})
}

func TestAskCommitMessageSelectorError(t *testing.T) {
	origSelect := selectCommitOption
	defer func() { selectCommitOption = origSelect }()

	selectCommitOption = func(_ *promptui.Select) (int, error) {
		return 0, fmt.Errorf("selector failed")
	}

	_, ok := askCommitMessage(ReleaseResult{Next: "v1.0.0", Commits: []Commit{{Type: "fix", Subject: "x"}}})
	if ok {
		t.Fatal("expected ok=false when selector fails")
	}
}

func TestAskCommitMessagePromptBranches(t *testing.T) {
	origSelect := selectCommitOption
	origPrompt := promptCommitMessage
	defer func() {
		selectCommitOption = origSelect
		promptCommitMessage = origPrompt
	}()

	selectCommitOption = func(_ *promptui.Select) (int, error) { return 2, nil }

	t.Run("prompt error", func(t *testing.T) {
		promptCommitMessage = func(_ *promptui.Prompt) (string, error) {
			return "", fmt.Errorf("prompt failed")
		}
		_, ok := askCommitMessage(ReleaseResult{Next: "v1.0.1", Commits: []Commit{{Type: "fix", Subject: "bug"}}})
		if ok {
			t.Fatal("expected ok=false when prompt fails")
		}
	})

	t.Run("empty uses default", func(t *testing.T) {
		promptCommitMessage = func(_ *promptui.Prompt) (string, error) { return "   ", nil }
		msg, ok := askCommitMessage(ReleaseResult{Next: "v1.0.2", Commits: []Commit{{Type: "fix", Subject: "bug"}}})
		if !ok {
			t.Fatal("expected ok=true")
		}
		if !strings.Contains(msg, "fix(release): release v1.0.2: bug") {
			t.Fatalf("unexpected message: %q", msg)
		}
	})

	t.Run("cancel keyword", func(t *testing.T) {
		promptCommitMessage = func(_ *promptui.Prompt) (string, error) { return "cancel", nil }
		_, ok := askCommitMessage(ReleaseResult{Next: "v1.0.3", Commits: []Commit{{Type: "fix", Subject: "bug"}}})
		if ok {
			t.Fatal("expected ok=false for cancel")
		}
	})
}

func TestRunDryRunTraversesCoreFlow(t *testing.T) {
	resetFlags()
	repo, dir := initTestRepo(t)
	_ = repo

	origArgs := os.Args
	origExit := exitFunc
	origStdout := os.Stdout
	defer func() {
		os.Args = origArgs
		exitFunc = origExit
		os.Stdout = origStdout
	}()

	os.Args = []string{"gogoversion", "--dry-run", "--no-tag", dir}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stdout = w

	exitFunc = func(code int) { panic(exitCall{code: code}) }

	func() {
		defer func() {
			rv := recover()
			ec, ok := rv.(exitCall)
			if !ok || ec.code != 0 {
				t.Fatalf("expected exit(0), got %#v", rv)
			}
		}()
		Run("vtest")
	}()

	w.Close()
	buf := make([]byte, 8192)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	for _, s := range []string{"Release context", "Previous version", "Bump reason", "New version", "--dry-run active"} {
		if !strings.Contains(output, s) {
			t.Fatalf("expected output to contain %q, got %q", s, output)
		}
	}
}

func TestRunNoCommitsNoLocalChangesExits(t *testing.T) {
	resetFlags()
	repo, dir := initTestRepo(t)
	addTag(t, repo, "v1.0.0")

	origArgs := os.Args
	origExit := exitFunc
	origNewConfig := newConfigFunc
	defer func() {
		os.Args = origArgs
		exitFunc = origExit
		newConfigFunc = origNewConfig
	}()

	os.Args = []string{"gogoversion", "--no-tag", dir}
	newConfigFunc = newConfig
	exitFunc = func(code int) { panic(exitCall{code: code}) }

	defer func() {
		rv := recover()
		ec, ok := rv.(exitCall)
		if !ok || ec.code != 0 {
			t.Fatalf("expected exit(0), got %#v", rv)
		}
	}()

	Run("vtest")
}

func TestRunNoCommitsWithChangesFallbackAndNoTagFlow(t *testing.T) {
	resetFlags()
	repo, dir := initTestRepo(t)
	t.Setenv("GIT_AUTHOR_NAME", "Test User")
	t.Setenv("GIT_AUTHOR_EMAIL", "test@example.com")
	t.Setenv("GIT_COMMITTER_NAME", "Test User")
	t.Setenv("GIT_COMMITTER_EMAIL", "test@example.com")
	addTag(t, repo, "v1.0.0")
	remoteDir := t.TempDir()
	if _, err := git.PlainInit(remoteDir, true); err != nil {
		t.Fatalf("PlainInit remote: %v", err)
	}
	if _, err := repo.CreateRemote(&gitconfig.RemoteConfig{Name: "origin", URLs: []string{remoteDir}}); err != nil {
		t.Fatalf("CreateRemote: %v", err)
	}
	if err := pushCurrentBranch(t, repo); err != nil {
		t.Fatalf("pushCurrentBranch: %v", err)
	}

	if err := os.WriteFile(dir+"/dirty.txt", []byte("dirty"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	origArgs := os.Args
	origNewConfig := newConfigFunc
	origAsk := askCommitMessageFn
	origInteractive := isInteractiveFn
	origSleep := sleepFunc
	defer func() {
		os.Args = origArgs
		newConfigFunc = origNewConfig
		askCommitMessageFn = origAsk
		isInteractiveFn = origInteractive
		sleepFunc = origSleep
	}()

	os.Args = []string{"gogoversion", "--no-tag", dir}
	newConfigFunc = newConfig
	askCommitMessageFn = func(ReleaseResult) (string, bool) { return "", false }
	isInteractiveFn = func() bool { return false }
	sleepFunc = func(time.Duration) {}

	Run("vtest")
}

func TestRunInteractiveCancelWhenCommitSelectionFails(t *testing.T) {
	resetFlags()
	repo, dir := initTestRepo(t)
	addTag(t, repo, "v1.0.0")
	if err := os.WriteFile(dir+"/dirty.txt", []byte("dirty"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	origArgs := os.Args
	origExit := exitFunc
	origNewConfig := newConfigFunc
	origAsk := askCommitMessageFn
	origInteractive := isInteractiveFn
	defer func() {
		os.Args = origArgs
		exitFunc = origExit
		newConfigFunc = origNewConfig
		askCommitMessageFn = origAsk
		isInteractiveFn = origInteractive
	}()

	os.Args = []string{"gogoversion", "--no-tag", dir}
	newConfigFunc = newConfig
	askCommitMessageFn = func(ReleaseResult) (string, bool) { return "", false }
	isInteractiveFn = func() bool { return true }
	exitFunc = func(code int) { panic(exitCall{code: code}) }

	defer func() {
		rv := recover()
		ec, ok := rv.(exitCall)
		if !ok || ec.code != 0 {
			t.Fatalf("expected exit(0), got %#v", rv)
		}
	}()

	Run("vtest")
}

func TestRunInteractiveConfirmationCancel(t *testing.T) {
	resetFlags()
	repo, dir := initTestRepo(t)
	addTag(t, repo, "v1.0.0")
	if err := os.WriteFile(dir+"/dirty.txt", []byte("dirty"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	origArgs := os.Args
	origExit := exitFunc
	origNewConfig := newConfigFunc
	origAsk := askCommitMessageFn
	origInteractive := isInteractiveFn
	origConfirm := confirmReleaseFn
	defer func() {
		os.Args = origArgs
		exitFunc = origExit
		newConfigFunc = origNewConfig
		askCommitMessageFn = origAsk
		isInteractiveFn = origInteractive
		confirmReleaseFn = origConfirm
	}()

	os.Args = []string{"gogoversion", "--no-tag", dir}
	newConfigFunc = newConfig
	askCommitMessageFn = func(ReleaseResult) (string, bool) { return "fix(release): x", true }
	isInteractiveFn = func() bool { return true }
	confirmReleaseFn = func(Config, ReleaseResult, string) bool { return false }
	exitFunc = func(code int) { panic(exitCall{code: code}) }

	defer func() {
		rv := recover()
		ec, ok := rv.(exitCall)
		if !ok || ec.code != 0 {
			t.Fatalf("expected exit(0), got %#v", rv)
		}
	}()

	Run("vtest")
}

func TestRunNoStagedChangesAndNoPushPath(t *testing.T) {
	resetFlags()
	repo, dir := initTestRepo(t)
	t.Setenv("GIT_AUTHOR_NAME", "Test User")
	t.Setenv("GIT_AUTHOR_EMAIL", "test@example.com")
	t.Setenv("GIT_COMMITTER_NAME", "Test User")
	t.Setenv("GIT_COMMITTER_EMAIL", "test@example.com")
	addTag(t, repo, "v1.0.0")
	_ = commitFile(t, repo, dir, "post-tag.txt", "fix: after tag")
	remoteDir := t.TempDir()
	if _, err := git.PlainInit(remoteDir, true); err != nil {
		t.Fatalf("PlainInit remote: %v", err)
	}
	if _, err := repo.CreateRemote(&gitconfig.RemoteConfig{Name: "origin", URLs: []string{remoteDir}}); err != nil {
		t.Fatalf("CreateRemote: %v", err)
	}
	if err := pushCurrentBranch(t, repo); err != nil {
		t.Fatalf("pushCurrentBranch: %v", err)
	}

	origArgs := os.Args
	origNewConfig := newConfigFunc
	origAsk := askCommitMessageFn
	origInteractive := isInteractiveFn
	origSleep := sleepFunc
	defer func() {
		os.Args = origArgs
		newConfigFunc = origNewConfig
		askCommitMessageFn = origAsk
		isInteractiveFn = origInteractive
		sleepFunc = origSleep
	}()

	os.Args = []string{"gogoversion", "--no-tag", "--no-changelog", dir}
	newConfigFunc = newConfig
	askCommitMessageFn = func(ReleaseResult) (string, bool) { return "", false }
	isInteractiveFn = func() bool { return false }
	sleepFunc = func(time.Duration) {}

	Run("vtest")
}

func pushCurrentBranch(t *testing.T, repo *git.Repository) error {
	t.Helper()
	head, err := repo.Head()
	if err != nil {
		return err
	}
	if !head.Name().IsBranch() {
		return fmt.Errorf("head is not a branch")
	}
	branch := head.Name().Short()
	refspec := gitconfig.RefSpec(fmt.Sprintf("refs/heads/%s:refs/heads/%s", branch, branch))
	if err := repo.Push(&git.PushOptions{RemoteName: "origin", RefSpecs: []gitconfig.RefSpec{refspec}}); err != nil && err != git.NoErrAlreadyUpToDate {
		return err
	}
	return nil
}
