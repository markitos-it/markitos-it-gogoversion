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
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
)

type configExitCall struct {
	code int
}

// resetFlags resets flag.CommandLine to a clean state between tests.
func resetFlags() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
}

func TestNewConfigShowHelp(t *testing.T) {
	resetFlags()
	os.Args = []string{"gogoversion", "--help"}

	cfg := newConfig()

	if !cfg.ShowHelp {
		t.Error("expected ShowHelp=true")
	}
}

func TestNewConfigShowVersion(t *testing.T) {
	resetFlags()
	os.Args = []string{"gogoversion", "--version"}

	cfg := newConfig()

	if !cfg.ShowVersion {
		t.Error("expected ShowVersion=true")
	}
}

func TestNewConfigDoubleDashShowsHelp(t *testing.T) {
	resetFlags()
	os.Args = []string{"gogoversion", "--"}

	cfg := newConfig()

	if !cfg.ShowHelp {
		t.Error("expected ShowHelp=true for '--' arg")
	}
}

func TestNewConfigNoArgsShowsHelp(t *testing.T) {
	resetFlags()
	os.Args = []string{"gogoversion"}

	cfg := newConfig()

	if !cfg.ShowHelp {
		t.Error("expected ShowHelp=true when no args are provided")
	}
}

func TestNewConfigDryRun(t *testing.T) {
	resetFlags()
	os.Args = []string{"gogoversion", "--dry-run", "."}

	cfg := newConfig()

	if !cfg.DryRun {
		t.Error("expected DryRun=true")
	}
	if cfg.RepoPath != "." {
		t.Errorf("RepoPath: got %q want %q", cfg.RepoPath, ".")
	}
}

func TestNewConfigNoTag(t *testing.T) {
	resetFlags()
	os.Args = []string{"gogoversion", "--no-tag", "."}

	cfg := newConfig()

	if !cfg.NoTag {
		t.Error("expected NoTag=true")
	}
}

func TestNewConfigNoChangelog(t *testing.T) {
	resetFlags()
	os.Args = []string{"gogoversion", "--no-changelog", "."}

	cfg := newConfig()

	if !cfg.NoChangelog {
		t.Error("expected NoChangelog=true")
	}
}

func TestNewConfigUndo(t *testing.T) {
	resetFlags()
	os.Args = []string{"gogoversion", "--undo", "."}

	cfg := newConfig()

	if !cfg.Undo {
		t.Error("expected Undo=true")
	}
	if cfg.RepoPath != "." {
		t.Errorf("RepoPath: got %q want %q", cfg.RepoPath, ".")
	}
}

func TestNewConfigRepoPath(t *testing.T) {
	resetFlags()
	os.Args = []string{"gogoversion", "/tmp/myrepo"}

	cfg := newConfig()

	if cfg.RepoPath != "/tmp/myrepo" {
		t.Errorf("RepoPath: got %q want %q", cfg.RepoPath, "/tmp/myrepo")
	}
}

func TestNewConfigUsageFunctionOutput(t *testing.T) {
	resetFlags()
	origArgs := os.Args
	defer func() { os.Args = origArgs }()

	os.Args = []string{"gogoversion", "--help"}
	_ = newConfig()

	var buf bytes.Buffer
	flag.CommandLine.SetOutput(&buf)
	flag.Usage()
	output := buf.String()

	for _, s := range []string{"Usage", "Notes", "Examples", "Flags", "--no-tag", "--no-changelog", "--undo"} {
		if !strings.Contains(output, s) {
			t.Fatalf("expected usage output to contain %q, got: %s", s, output)
		}
	}
}

func TestColorizerWithoutANSI(t *testing.T) {
	var buf bytes.Buffer
	paint := colorizer(&buf)
	if got := paint("hello", ansiBold); got != "hello" {
		t.Fatalf("got %q want %q", got, "hello")
	}
}

func TestSupportsANSIStatError(t *testing.T) {
	file, err := os.CreateTemp(t.TempDir(), "staterr-*.tmp")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	_ = file.Close()

	if got := supportsANSI(file); got {
		t.Fatal("expected supportsANSI=false when file.Stat() fails")
	}
}

func TestNewConfigExitPaths(t *testing.T) {
	testCases := []struct {
		name string
		args []string
	}{
		{name: "invalid flag parse", args: []string{"gogoversion", "--invalid"}},
		{name: "undo without path", args: []string{"gogoversion", "--undo"}},
		{name: "too many args", args: []string{"gogoversion", ".", "extra"}},
		{name: "invalid path dash", args: []string{"gogoversion", "-"}},
		{name: "path not last", args: []string{"gogoversion", ".", "--no-tag"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resetFlags()
			origArgs := os.Args
			origExit := configExitFunc
			origStderr := os.Stderr
			defer func() {
				os.Args = origArgs
				configExitFunc = origExit
				os.Stderr = origStderr
			}()

			r, w, err := os.Pipe()
			if err != nil {
				t.Fatalf("os.Pipe: %v", err)
			}
			os.Stderr = w

			os.Args = tc.args
			configExitFunc = func(code int) { panic(configExitCall{code: code}) }

			defer func() {
				w.Close()
				buf := make([]byte, 1024)
				n, _ := r.Read(buf)
				_ = string(buf[:n])

				rv := recover()
				ec, ok := rv.(configExitCall)
				if !ok {
					t.Fatalf("expected configExitCall panic, got %#v", rv)
				}
				if ec.code != 2 {
					t.Fatalf("expected exit code 2, got %d", ec.code)
				}
			}()

			_ = newConfig()
		})
	}
}

func TestColorizerWithANSI(t *testing.T) {
	origCheck := supportsANSICheck
	defer func() { supportsANSICheck = origCheck }()
	supportsANSICheck = func(_ io.Writer) bool { return true }

	var buf bytes.Buffer
	paint := colorizer(&buf)
	got := paint("hello", ansiBold)
	want := fmt.Sprintf("%shello%s", ansiBold, ansiReset)
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
