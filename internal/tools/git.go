package tools

import (
	"errors"
	"os"

	"github.com/go-git/go-git/v6"
)

func CloneOrUpdateRepo(repoPath, repoURL string) error {
	info, err := os.Stat(repoPath)
	if err == nil {
		if !info.IsDir() {
			if err := os.RemoveAll(repoPath); err != nil {
				return err
			}
			_, err = git.PlainClone(repoPath, &git.CloneOptions{
				URL: repoURL,
			})
			return err
		}

		repo, err := git.PlainOpen(repoPath)
		if err != nil {
			if err := os.RemoveAll(repoPath); err != nil {
				return err
			}
			_, err = git.PlainClone(repoPath, &git.CloneOptions{
				URL: repoURL,
			})
			return err
		}

		worktree, err := repo.Worktree()
		if err != nil {
			return err
		}

		err = worktree.Pull(&git.PullOptions{
			RemoteName: "origin",
		})
		if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
			return err
		}
		return nil
	}

	if !os.IsNotExist(err) {
		return err
	}

	_, err = git.PlainClone(repoPath, &git.CloneOptions{
		URL: repoURL,
	})
	return err
}
