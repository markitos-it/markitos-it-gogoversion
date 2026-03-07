package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInstallAliases_CreatesSymlinksForGogoversion(t *testing.T) {
	tmpDir := t.TempDir()
	fakeBinary := filepath.Join(tmpDir, "gogoversion")
	if err := os.WriteFile(fakeBinary, []byte("fake"), 0755); err != nil {
		t.Fatal(err)
	}

	origExec := executableFunc
	origSymlink := symlinkFunc
	origRemove := removeFunc
	origExit := exitFunc
	t.Cleanup(func() {
		executableFunc = origExec
		symlinkFunc = origSymlink
		removeFunc = origRemove
		exitFunc = origExit
	})

	executableFunc = func() (string, error) { return fakeBinary, nil }

	created := map[string]string{}
	symlinkFunc = func(oldname, newname string) error {
		created[filepath.Base(newname)] = oldname
		return nil
	}
	removeFunc = func(name string) error { return nil }
	exitFunc = func(code int) {}

	installAliases()

	if _, ok := created["ggv"]; !ok {
		t.Error("expected alias 'ggv' to be created")
	}
	if _, ok := created["gogov"]; !ok {
		t.Error("expected alias 'gogov' to be created")
	}
	if _, ok := created["gogoversion"]; ok {
		t.Error("alias 'gogoversion' should NOT be created when binary is already named 'gogoversion'")
	}
}

func TestInstallAliases_CreatesGogoversionAliasForPlatformBinary(t *testing.T) {
	tmpDir := t.TempDir()
	fakeBinary := filepath.Join(tmpDir, "gogoversion-linux-amd64")
	if err := os.WriteFile(fakeBinary, []byte("fake"), 0755); err != nil {
		t.Fatal(err)
	}

	origExec := executableFunc
	origSymlink := symlinkFunc
	origRemove := removeFunc
	origExit := exitFunc
	t.Cleanup(func() {
		executableFunc = origExec
		symlinkFunc = origSymlink
		removeFunc = origRemove
		exitFunc = origExit
	})

	executableFunc = func() (string, error) { return fakeBinary, nil }

	created := map[string]string{}
	symlinkFunc = func(oldname, newname string) error {
		created[filepath.Base(newname)] = oldname
		return nil
	}
	removeFunc = func(name string) error { return nil }
	exitFunc = func(code int) {}

	installAliases()

	if _, ok := created["ggv"]; !ok {
		t.Error("expected alias 'ggv' to be created")
	}
	if _, ok := created["gogov"]; !ok {
		t.Error("expected alias 'gogov' to be created")
	}
	if _, ok := created["gogoversion"]; !ok {
		t.Error("expected alias 'gogoversion' to be created for platform binary")
	}
}

func TestUninstallAliases_RemovesSymlinks(t *testing.T) {
	tmpDir := t.TempDir()
	fakeBinary := filepath.Join(tmpDir, "gogoversion")
	if err := os.WriteFile(fakeBinary, []byte("fake"), 0755); err != nil {
		t.Fatal(err)
	}

	origExec := executableFunc
	origSymlink := symlinkFunc
	origRemove := removeFunc
	origExit := exitFunc
	t.Cleanup(func() {
		executableFunc = origExec
		symlinkFunc = origSymlink
		removeFunc = origRemove
		exitFunc = origExit
	})

	executableFunc = func() (string, error) { return fakeBinary, nil }
	symlinkFunc = func(oldname, newname string) error { return nil }

	removed := []string{}
	removeFunc = func(name string) error {
		removed = append(removed, filepath.Base(name))
		return nil
	}
	exitFunc = func(code int) {}

	uninstallAliases()

	if len(removed) != 2 {
		t.Errorf("expected 2 aliases removed, got %d: %v", len(removed), removed)
	}
	for _, alias := range []string{"ggv", "gogov"} {
		found := false
		for _, r := range removed {
			if r == alias {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected alias %q to be removed", alias)
		}
	}
}
