package main

import (
	"fmt"
	"log"
	"net/http"

	"os"

	"github.com/google/go-github/github"
	"github.com/satori/go.uuid"
	"os/exec"
	"strings"
)

var eventChan chan *github.PullRequestEvent

const rootDir = "/var/www/github-bot"
const botUsername = "vektorbot"

type GitHubEventMonitor struct {
	secret []byte
}

func (s *GitHubEventMonitor) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	payload, err := github.ValidatePayload(r, s.secret)
	if err != nil {
		log.Fatal(err)
	}
	event, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		log.Fatal(err)
	}
	switch event := event.(type) {
	case *github.PullRequestEvent:
		if *event.Sender.Login == botUsername {
			break
		}

		if *event.Action == "opened" || *event.Action == "synchronize" {
			go func(event *github.PullRequestEvent) {
				eventChan <- event
			}(event)
		}

	}
}

func doBotStuff(event *github.PullRequestEvent) {
	branch := *event.PullRequest.Head.Ref
	repoFolderName, err := cloneRepo(*event.Repo.SSHURL)
	if err != nil {
		fmt.Println("Failed to clone repository " + *event.Repo.SSHURL)
		return
	}
	defer cleanUp(repoFolderName)

	if err := createParametersFile(repoFolderName); err != nil {
		fmt.Println("Failed to create parameters file")
		return
	}

	if err := installDependencies(); err != nil {
		fmt.Println("Failed to install dependencies")
		return
	}
	if err := createDatabase(); err != nil {
		fmt.Println("Failed to create database")
		return
	}
	if err := checkout(branch); err != nil {
		fmt.Println("Failed checkout " + branch)
		return
	}
	if err := installDependencies(); err != nil {
		fmt.Println("Failed to install dependencies")
		return
	}
	if err := executeMigrations(); err != nil {
		fmt.Println("Failed to execute migrations")
		return
	}
	if err := createMigration(); err != nil {
		fmt.Println("Failed to create migration")
		return
	}
	if err := fixCodeStyle(); err != nil {
		fmt.Println("Failed to fix code style")
		return
	}
	if err := gitPush(); err != nil {
		fmt.Println("Failed to git push")
		return
	}
}

func cloneRepo(repo string) (string, error) {
	repoFolderName := fmt.Sprintf("%s", uuid.NewV4())
	if _, err := executeCommand(fmt.Sprintf("git clone %s %s/%s", repo, rootDir, repoFolderName)); err != nil {
		return "", err
	}

	if err := cd(fmt.Sprintf("%s/%s", rootDir, repoFolderName)); err != nil {
		return "", err
	}

	return repoFolderName, nil
}

func installDependencies() error {
	_, err := executeCommand("composer install -n")
	return err
}

func createParametersFile(dbName string) error {
	if _, err := executeCommand(fmt.Sprintf("cp %s/parameters.yml app/config/parameters.yml", rootDir)); err != nil {
		return err
	}
	_, err := executeCommand(fmt.Sprintf("sed -i 's/dbname/%s/g' app/config/parameters.yml", dbName))
	return err
}

func createDatabase() error {
	if _, err := executeCommand("php app/console doctrine:database:create"); err != nil {
		return err
	}
	if _, err := executeCommand("php app/console doctrine:schema:create"); err != nil {
		return err
	}
	_, err := executeCommand("php app/console doctrine:migrations:version --add --all -n")
	return err
}

func checkout(branch string) error {
	_, err := executeCommand(fmt.Sprintf("git checkout %s", branch))
	return err
}

func executeMigrations() error {
	_, err := executeCommand("php app/console doctrine:migrations:migrate -n")
	if err != nil {
		return err
	}
	return nil
}

func createMigration() error {
	migrationResult, err := executeCommand("php app/console doctrine:migrations:diff")
	if err != nil {
		return err
	}

	if strings.Contains(migrationResult, "Generated new migration") {
		_, err = executeCommand("git add app/DoctrineMigrations/")
		if err != nil {
			return err
		}

		_, err = executeCommand("git commit -m 'Create database migration'")
		if err != nil {
			return err
		}
	}

	return nil
}

func fixCodeStyle() error {
	codeStyleResult, err := executeCommand("npm run cs")
	if err != nil {
		return err
	}
	if strings.Contains(codeStyleResult, "1)") {
		if _, err = executeCommand("git add src/"); err != nil {
			return err
		}
		if _, err = executeCommand("git commit -m 'Fix code style'"); err != nil {
			return err
		}
	}

	return nil
}

func gitPush() error {
	_, err := executeCommand("git push")
	return err
}

func cleanUp(folderName string) error {
	if _, err := executeCommand("php app/console doctrine:database:drop --force"); err != nil {
		return err
	}
	if len(folderName) > 1 {
		if _, err := executeCommand(fmt.Sprintf("rm -rf %s/%s/", rootDir, folderName)); err != nil {
			return err
		}
	}
	return cd(rootDir)
}

func executeCommand(command string) (string, error) {
	cmd := exec.Command("sh", "-c", command)
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Failed to execute " + command)
		return "", err
	}

	return fmt.Sprintf("%s", output), nil
}

func cd(dir string) error {
	return os.Chdir(dir)
}

func main() {
	eventChan = make(chan *github.PullRequestEvent)
	go func() {
		for event := range eventChan {
			doBotStuff(event)
		}
	}()

	secret := os.Getenv("GITHUB_WEBHOOKS_SECRET")
	eventMonitor := GitHubEventMonitor{secret: []byte(secret)}

	http.HandleFunc("/webhooks", eventMonitor.ServeHTTP)
	fmt.Println("Listening to webhooks on port 5555")
	log.Fatal(http.ListenAndServe(":5555", nil))
}
