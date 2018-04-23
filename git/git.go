package gitclient

import (
	"fmt"
	"github.com/satori/go.uuid"
	"github.com/vektorprogrammet/github-bot/cmd"
	"strings"
)

func CloneRepo(repo, rootDir string) (string, error) {
	repoFolderName := fmt.Sprintf("%s", uuid.NewV4())
	if _, err := cmd.Execute(fmt.Sprintf("git clone %s %s/%s", repo, rootDir, repoFolderName), rootDir); err != nil {
		return "", err
	}

	return repoFolderName, nil
}

func Checkout(branch, workingDirectory string) error {
	_, err := cmd.Execute(fmt.Sprintf("git checkout %s", branch), workingDirectory)
	return err
}

func CheckoutNewBranch(branch, workingDirectory string) error {
	_, err := cmd.Execute(fmt.Sprintf("git checkout -b %s", branch), workingDirectory)
	return err
}

func Push(workingDirectory string) error {
	_, err := cmd.Execute("git push", workingDirectory)
	return err
}

func PushToBranch(workingDirectory, branch string) error {
	_, err := cmd.Execute("git push -u origin " + branch, workingDirectory)
	return err
}

func PullRebase(workingDirectory string) error {
	_, err := cmd.Execute("git pull --rebase", workingDirectory)
	return err
}

func Commit(message, workingDirectory string) (bool, error) {
	output, err := cmd.Execute(fmt.Sprintf("git commit -m '%s'", message), workingDirectory)
	if err != nil {
		return false, err
	}
	return !strings.Contains(output, "nothing to commit"), nil
}

func Add(path, workingDirectory string) error {
	_, err := cmd.Execute(fmt.Sprintf("git add %s", path), workingDirectory)
	return err
}