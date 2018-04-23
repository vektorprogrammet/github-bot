package bot

import (
	"fmt"
	"github.com/vektorprogrammet/github-bot/symfony"
	"github.com/vektorprogrammet/github-bot/git"
	"github.com/google/go-github/github"
	"github.com/vektorprogrammet/github-bot/cmd"
)

type DbMigrationBot struct{
	Bot
	symfony.Symfony
}

func (bot DbMigrationBot) HandleEvent(event interface{}) {
	fmt.Println("DbMigrationBot handling event")
	e, ok := event.(*github.PullRequestEvent)
	if !ok {
		fmt.Println("Not a pull request event")
		return
	}
	if *e.Sender.Login == bot.Username {
		fmt.Println("Event from bot")
		return
	}

	if *e.Action != "opened" && *e.Action != "synchronize" {
		fmt.Println("Action was not opened or synchronize")
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

	if err := bot.Symfony.CreateParametersFile(repoFolderName); err != nil {
		fmt.Println("Failed to create parameters file")
		return
	}

	if err := bot.Symfony.InstallDependencies(); err != nil {
		fmt.Println("Failed to install dependencies")
		return
	}
	if err := bot.Symfony.CreateDatabase(); err != nil {
		fmt.Println("Failed to create database")
		return
	}
	//if err := bot.Symfony.DropDatabase(); err != nil {
	//	fmt.Println("Failed to drop database")
	//	return
	//}
	if err := gitclient.Checkout(branch, bot.Symfony.WorkingDirectory); err != nil {
		fmt.Println("Failed checkout " + branch)
		return
	}
	if err := bot.Symfony.InstallDependencies(); err != nil {
		fmt.Println("Failed to install dependencies")
		return
	}
	if err := bot.Symfony.ExecuteMigrations(); err != nil {
		fmt.Println("Failed to execute migrations")
		return
	}
	migrationCreated, err := bot.Symfony.CreateMigration()
	if err != nil {
		fmt.Println("Failed to create migration")
		return
	}

	if !migrationCreated {
		return
	}

	if err := gitclient.Add("app/DoctrineMigrations/", bot.Symfony.WorkingDirectory); err != nil {
		fmt.Println("Failed to git add")
		return
	}

	committed, err := gitclient.Commit("Create database migration", bot.Symfony.WorkingDirectory)
	if err != nil {
		fmt.Println("Failed to git commit")
		return
	}
	if !committed {
		return
	}
	if err := gitclient.Push(bot.Symfony.WorkingDirectory); err != nil {
		fmt.Println("Failed to git push")
		return
	}
}

func (bot DbMigrationBot) cleanUp(folderName string) error {
	defer func() {
		if len(folderName) > 1 {
			cmd.RemoveFolder(fmt.Sprintf("%s/%s/", bot.RootDir, folderName))
		}
	}()

	if err := bot.Symfony.DropDatabase(); err != nil {
		return err
	}

	return nil
}
