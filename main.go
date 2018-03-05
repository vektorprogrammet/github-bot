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
		if *event.Sender.Login == "vektorbot" {
			break
		}

		go func(event *github.PullRequestEvent) {
			eventChan <- event
		}(event)
	}
}

func doBotStuff(event *github.PullRequestEvent) {
	branch := *event.PullRequest.Head.Ref
	repoFolderName := cloneRepo(*event.Repo.SSHURL)
	defer cleanUp(repoFolderName)

	createParametersFile(repoFolderName)
	installDependencies()
	createDatabase()
	checkout(branch)
	executeMigrations()
	createMigration()
	fixCodeStyle()
	gitPush()
}

func cloneRepo(repo string) string {
	repoFolderName := fmt.Sprintf("%s", uuid.NewV4())
	executeCommand(fmt.Sprintf("git clone %s /var/www/github-bot/%s", repo, repoFolderName))
	cd(fmt.Sprintf("/var/www/github-bot/%s", repoFolderName))

	return repoFolderName
}

func installDependencies() {
	executeCommand("composer install -n")
}

func createParametersFile(dbName string) {
	executeCommand("cp /var/www/github-bot/parameters.yml app/config/parameters.yml")
	executeCommand(fmt.Sprintf("sed -i 's/dbname/%s/g' app/config/parameters.yml", dbName))
}

func createDatabase() {
	executeCommand("php app/console doctrine:database:create")
	executeCommand("php app/console doctrine:schema:create")
	executeCommand("php app/console doctrine:migrations:version --add --all -n")
}

func checkout(branch string) {
	executeCommand(fmt.Sprintf("git checkout %s", branch))
}

func executeMigrations() {
	executeCommand("php app/console doctrine:migrations:migrate -n")
}

func createMigration() {
	migrationResult := executeCommand("php app/console doctrine:migrations:diff")
	if strings.Contains(migrationResult, "Generated new migration") {
		executeCommand("git add app/DoctrineMigrations/")
		executeCommand("git commit -m 'Create database migration'")
	}
}

func fixCodeStyle() {
	codeStyleResult := executeCommand("npm run cs")
	if strings.Contains(codeStyleResult, "1)") {
		executeCommand("git add src/")
		executeCommand("git commit -m 'Fix code style'")
	}
}

func gitPush() {
	executeCommand("git push")
}

func cleanUp(folderName string) {
	executeCommand("php app/console doctrine:database:drop --force")
	if len(folderName) > 1 {
		executeCommand(fmt.Sprintf("rm -rf /var/www/github-bot/%s/", folderName))
	}
	cd("/var/www/github-bot")
}

func executeCommand(command string) string {
	//fmt.Println(command)
	cmd := exec.Command("sh", "-c", command)
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Failed to execute " + command)
		fmt.Println(err)
	}
	//fmt.Printf("%s", output)

	return fmt.Sprintf("%s", output)
}

func cd(dir string) {
	os.Chdir(dir)
}

func main() {
	eventChan = make(chan *github.PullRequestEvent)

	go func(){
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