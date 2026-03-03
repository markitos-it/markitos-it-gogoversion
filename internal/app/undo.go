package app

import "fmt"

func undoLastRelease(repoPath string) error {
	repo, err := openRepository(repoPath)
	if err != nil {
		return err
	}

	tagName, err := latestTagName(repo)
	if err != nil {
		return err
	}
	if tagName == "" {
		fmt.Println("⊘  No semver tags found to undo.")
		return nil
	}

	if err := deleteTag(repo, tagName); err != nil {
		return err
	}
	fmt.Printf("✔  Tag %s removed\n", tagName)

	removed, err := removeChangelogEntry(repoPath, tagName)
	if err != nil {
		return err
	}
	if removed {
		fmt.Printf("✔  Entry %s removed from CHANGELOG.md\n", tagName)
	} else {
		fmt.Println("ℹ  No matching entry found in CHANGELOG.md")
	}

	fmt.Printf("\n↩  Undo of %s completed\n", tagName)
	return nil
}
