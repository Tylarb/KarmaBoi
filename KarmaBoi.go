package main

import (
	"fmt"
	"log"
	"os"

	"github.com/nlopes/slack"
)

// Get bot name and token from env
var slackBotToken = os.Getenv("SLACK_BOT_TOKEN")
var slackBotName = os.Getenv("SLACK_BOT_NAME")

func main() {
	fmt.Println("slack token:", slackBotToken)
	fmt.Println("slack name:", slackBotName)

	fmt.Println("starting bot")
	api := slack.New(slackBotToken)
	users, err := api.GetUsers()
	if err != nil {
		log.Fatal(err)
	}
	for _, user := range users {
		fmt.Printf("name: %s, id: %s\n", user.Name, user.ID)
		if user.Name == slackBotName {
			fmt.Printf("Found bot ID %s for name %s\n", user.ID, user.Name)
		}
	}

}
