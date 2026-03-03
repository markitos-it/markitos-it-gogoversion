package app

import "fmt"

var (
	undoLatestTagNameFn = latestTagName
	undoDeleteTagFn     = deleteTag
)

func undoLastRelease(repoPath string) error {
	repo, err := openRepository(repoPath)
	if err != nil {
		return err
	}

	tagName, err := undoLatestTagNameFn(repo)
	if err != nil {
		return err
	}
	if tagName == "" {
		fmt.Println("⊘  No semver tags found to undo.")
		return nil
	}

	if err := undoDeleteTagFn(repo, tagName); err != nil {
		return err
	}
	fmt.Printf("✔  Tag %s removed\n", tagName)
	if err := deleteRemoteTag(repo, tagName); err == nil {
		fmt.Printf("✔  Remote tag %s removed from origin\n", tagName)
	} else {
		fmt.Printf("ℹ  Remote tag %s was not removed from origin: %v\n", tagName, err)
	}

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
