package bot

import (
	"fmt"
	"github.com/vektorprogrammet/github-bot/symfony"
	"github.com/vektorprogrammet/github-bot/git"
	"github.com/vektorprogrammet/github-bot/cmd"
	"time"
)

type ComposerUpdateBot struct{
	Bot
	symfony.Symfony
}

func (bot ComposerUpdateBot) UpdatePackages(repoUrl string) {
	fmt.Println("Updating packages")
	repoFolderName, err := gitclient.CloneRepo(repoUrl, bot.RootDir)
	if err != nil {
		fmt.Println("Failed to clone repository " + repoUrl)
		return
	}
	bot.Symfony = symfony.Symfony{
		WorkingDirectory: fmt.Sprintf("%s/%s", bot.RootDir, repoFolderName),
	}
	defer bot.cleanUp(repoFolderName)

	t := time.Now()
	branch := "dependencies/" + t.Format("2006-01-02_15-04-05")

	if err := gitclient.CheckoutNewBranch(branch, bot.Symfony.WorkingDirectory); err != nil {
		fmt.Println("Failed checkout " + branch)
		return
	}
	if err := bot.Symfony.InstallDependencies(); err != nil {
		fmt.Println("Failed to install dependencies")
		return
	}

	updated, err := bot.Symfony.UpdateDependencies()
	if err != nil {
		fmt.Println("Failed to update dependencies")
		return
	}

	if !updated {
		return
	}

	if err := gitclient.Add("composer.lock", bot.Symfony.WorkingDirectory); err != nil {
		fmt.Println("Failed to git add")
		return
	}
	committed, err := gitclient.Commit("Update dependencies", bot.Symfony.WorkingDirectory)
	if err != nil {
		fmt.Println("Failed to git commit")
		return
	}
	if !committed {
		return
	}
	if err := gitclient.PushToBranch(bot.Symfony.WorkingDirectory, branch); err != nil {
		fmt.Println("Failed to git push")
		return
	}
}

func (bot ComposerUpdateBot) cleanUp(folderName string) error {
	if len(folderName) > 1 {
		if err := cmd.RemoveFolder(fmt.Sprintf("%s/%s/", bot.RootDir, folderName)); err != nil {
			return err
		}
	}
	return nil
}
