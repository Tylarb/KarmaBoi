/*
Karma tracking bot with additional features. Integrates with Slack, runs on PCF


Released under MIT license, copyright 2018 Tyler Ramer

*/

package main

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/nlopes/slack"
)

// Get bot name and token from env
var slackBotToken = os.Getenv("SLACK_BOT_TOKEN")
var slackBotName = os.Getenv("SLACK_BOT_NAME")

// log levels, see logrus docs
func init() {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
}

func getBotID(botName string, sc *slack.Client) (botID string) {
	users, err := sc.GetUsers()
	if err != nil {
		log.Fatal(err)
	}
	for _, user := range users {
		if user.Name == botName {
			log.WithFields(log.Fields{"ID": user.ID, "name": user.Name}).Debug("Found bot:")
			botID = user.ID
		}
	}
	return
}

func main() {
	sc := slack.New(slackBotToken)
	botID := getBotID(slackBotName, sc)
	rtm := sc.NewRTM()
	go rtm.ManageConnection()
	log.Info("Connected to slack")

	for slackEvent := range rtm.IncomingEvents {
		switch ev := slackEvent.Data.(type) {
		case *slack.HelloEvent:
			// Ignored
		case *slack.MessageEvent:
			channel := ev.Channel
			log.WithField(channel, channel).Debug("message on channel:")
			out := fmt.Sprintf("Hello my name is %s", botID)
			rtm.SendMessage(rtm.NewOutgoingMessage(out, channel))

		}

	}

}
