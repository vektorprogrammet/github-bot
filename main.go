package main

import (
	"fmt"
	"log"
	"net/http"

	"os"
	"github.com/robfig/cron"
	"github.com/google/go-github/github"
	"github.com/vektorprogrammet/github-bot/bot"
)

var eventChan chan interface{}

const rootDir = "/var/www/github-bot"
const botUsername = "vektorbot"

type GitHubEventMonitor struct {
	secret []byte
}

type GitHubEventHandler interface {
	HandleEvent(event interface{})
}

func (s *GitHubEventMonitor) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	payload, err := github.ValidatePayload(r, s.secret)
	if err != nil {
		fmt.Printf("Failed to valdidate payload: %s\n", err)
		return
	}
	event, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		fmt.Printf("Failed to parse webhook: %s\n", err)
		return
	}
	go func(event interface{}) {
		eventChan <- event
	}(event)
}

func startGitHubEventBots() {
	gitHubBot := bot.Bot{
		Username: botUsername,
		RootDir:  rootDir,
	}
	gitHubEventBots := []GitHubEventHandler{
		bot.CodeStyleBot{
			Bot: gitHubBot,
		},
		bot.DbMigrationBot{
			Bot: gitHubBot,
		},
	}
	eventChan = make(chan interface{})
	go func() {
		for event := range eventChan {
			fmt.Println("Event received")
			for _, gitHubBot := range gitHubEventBots {
				go gitHubBot.HandleEvent(event)
			}
		}
	}()
}

func startUpdateBot(repoUrl, cronInterval string) {
	composerUpdateBot := bot.ComposerUpdateBot{
		Bot: bot.Bot{
			Username: botUsername,
			RootDir:  rootDir,
		},
	}
	c := cron.New()
	c.AddFunc(cronInterval, func() {
		composerUpdateBot.UpdatePackages(repoUrl)
	})
	c.Start()
}

func main() {
	startGitHubEventBots()
	startUpdateBot("git@github.com:vektorprogrammet/vektorprogrammet.git", "0 0 20 * * SUN")

	secret := os.Getenv("GITHUB_WEBHOOKS_SECRET")
	eventMonitor := GitHubEventMonitor{secret: []byte(secret)}

	http.HandleFunc("/webhooks", eventMonitor.ServeHTTP)
	fmt.Println("Listening to webhooks on port 5555")
	log.Fatal(http.ListenAndServe(":5555", nil))
}
