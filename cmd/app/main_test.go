package main

import "testing"

func TestMainCallsRunWithVersion(t *testing.T) {
	originalRun := run
	originalVersion := version
	t.Cleanup(func() {
		run = originalRun
		version = originalVersion
	})

	version = "v9.9.9-test"
	called := false
	gotVersion := ""
	run = func(v string) {
		called = true
		gotVersion = v
	}

	main()

	if !called {
		t.Fatal("expected run to be called")
	}
	if gotVersion != version {
		t.Fatalf("expected version %q, got %q", version, gotVersion)
	}
}
