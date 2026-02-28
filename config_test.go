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
	"flag"
	"os"
	"testing"
)

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
