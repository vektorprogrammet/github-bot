package bot

import (
	"fmt"
	"github.com/vektorprogrammet/github-bot/symfony"
	"github.com/vektorprogrammet/github-bot/git"
	"github.com/google/go-github/github"
	"github.com/vektorprogrammet/github-bot/cmd"
)

type CodeStyleBot struct{
	Bot
	symfony.Symfony
}

func (bot CodeStyleBot) HandleEvent(event interface{}) {
	fmt.Println("CodeStyle handling event")
	e, ok := event.(*github.PullRequestEvent)
	if !ok {
		fmt.Println("Not a pull request event")
		return
	}
	if *e.Sender.Login == bot.Username {
		return
	}

	if *e.Action != "opened" && *e.Action != "synchronize" {
		return
	}

	branch := *e.PullRequest.Head.Ref
	repoFolderName, err := gitclient.CloneRepo(*e.Repo.SSHURL, bot.RootDir)
	if err != nil {
		fmt.Println("Failed to clone repository " + *e.Repo.SSHURL)
		return
	}
	bot.Symfony = symfony.Symfony{
		WorkingDirectory: fmt.Sprintf("%s/%s", bot.RootDir, repoFolderName),
	}
	defer bot.cleanUp(repoFolderName)

	if err := bot.Symfony.InstallDependencies(); err != nil {
		fmt.Println("Failed to install dependencies")
		return
	}
	if err := gitclient.Checkout(branch, bot.Symfony.WorkingDirectory); err != nil {
		fmt.Println("Failed checkout " + branch)
		return
	}
	fixed, err := bot.Symfony.FixCodeStyle()
	if err != nil {
		fmt.Println("Failed to fix code style")
		return
	}

	if !fixed {
		return
	}

	if err := gitclient.Add("src/", bot.Symfony.WorkingDirectory); err != nil {
		fmt.Println("Failed to git add")
		return
	}
	committed, err := gitclient.Commit("Fix code style", bot.Symfony.WorkingDirectory)
	if err != nil {
		fmt.Println("Failed to git commit")
		return
	}
	if !committed {
		return
	}
	if err := gitclient.PullRebase(bot.Symfony.WorkingDirectory); err != nil {
		fmt.Println("Failed to git pull --rebase")
		return
	}
	if err := gitclient.Push(bot.Symfony.WorkingDirectory); err != nil {
		fmt.Println("Failed to git push")
		return
	}
}

func (bot CodeStyleBot) cleanUp(folderName string) error {
	if len(folderName) > 1 {
		if err := cmd.RemoveFolder(fmt.Sprintf("%s/%s/", bot.RootDir, folderName)); err != nil {
			return err
		}
	}
	return nil
}
