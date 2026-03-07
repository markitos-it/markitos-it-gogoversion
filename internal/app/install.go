package app

import (
	"fmt"
	"os"
	"path/filepath"
)

var (
	executableFunc = os.Executable
	symlinkFunc    = os.Symlink
	removeFunc     = os.Remove
)

var allAliases = []string{"ggv", "gogov", "gogoversion"}

func resolveAliasesForSelf(selfName string) []string {
	if selfName == "gogoversion" {
		return []string{"ggv", "gogov"}
	}
	return allAliases
}

// resolveSelf returns the real path directory and base name of the running executable.
func resolveSelf() (dir, selfName string, err error) {
	self, err := executableFunc()
	if err != nil {
		return "", "", fmt.Errorf("cannot determine executable path: %w", err)
	}
	self, err = filepath.EvalSymlinks(self)
	if err != nil {
		return "", "", fmt.Errorf("cannot resolve symlinks: %w", err)
	}
	return filepath.Dir(self), filepath.Base(self), nil
}

func installAliases() {
	dir, selfName, err := resolveSelf()
	if err != nil {
		fmt.Fprintf(os.Stderr, "✖  %v\n", err)
		exitFunc(1)
		return
	}

	self := filepath.Join(dir, selfName)
	aliases := resolveAliasesForSelf(selfName)

	fmt.Printf("✨ Installing aliases for %s in %s\n\n", selfName, dir)
	for _, alias := range aliases {
		target := filepath.Join(dir, alias)
		if err := removeFunc(target); err != nil && !os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "✖  Cannot remove existing alias %s: %v\n", alias, err)
			continue
		}
		if err := symlinkFunc(self, target); err != nil {
			fmt.Fprintf(os.Stderr, "✖  %-20s → %v\n", alias, err)
		} else {
			fmt.Printf("✓  %-20s → %s\n", alias, self)
		}
	}
	fmt.Println()
}

func uninstallAliases() {
	dir, selfName, err := resolveSelf()
	if err != nil {
		fmt.Fprintf(os.Stderr, "✖  %v\n", err)
		exitFunc(1)
		return
	}

	aliases := resolveAliasesForSelf(selfName)

	fmt.Printf("🗑  Uninstalling aliases for %s from %s\n\n", selfName, dir)
	for _, alias := range aliases {
		target := filepath.Join(dir, alias)
		err := removeFunc(target)
		if err != nil && !os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "✖  %-20s → %v\n", alias, err)
		} else if err == nil {
			fmt.Printf("✓  removed %-14s (%s)\n", alias, target)
		} else {
			fmt.Printf("–  %-20s (not found, skipped)\n", alias)
		}
	}
	fmt.Println()
}
