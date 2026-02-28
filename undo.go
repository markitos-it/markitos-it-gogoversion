package main

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
		fmt.Println("⊘  No hay tags semver para deshacer.")
		return nil
	}

	if err := deleteTag(repo, tagName); err != nil {
		return err
	}
	fmt.Printf("✔  Tag %s eliminado\n", tagName)

	removed, err := removeChangelogEntry(repoPath, tagName)
	if err != nil {
		return err
	}
	if removed {
		fmt.Printf("✔  Entrada %s eliminada de CHANGELOG.md\n", tagName)
	} else {
		fmt.Println("ℹ  No se encontró entrada correspondiente en CHANGELOG.md")
	}

	fmt.Printf("\n↩  Undo de %s completado\n", tagName)
	return nil
}
