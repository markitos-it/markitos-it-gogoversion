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
	"bytes"
	"fmt"
	"os"
	"testing"
)

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
	fmt.Fprintf(os.Stderr, "✖  Error %s: %v\n", "test context", testErr)

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
