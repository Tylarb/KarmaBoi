/*
Karma tracking bot with additional features. Integrates with Slack, runs on PCF


Released under MIT license, copyright 2018 Tyler Ramer

*/

package main

import (
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
	log.WithField("ID", botID).Debug("Bot ID returned")
	rtm := sc.NewRTM()
	go rtm.ManageConnection()
	log.Info("Connected to slack")

	for slackEvent := range rtm.IncomingEvents {
		switch ev := slackEvent.Data.(type) {
		case *slack.HelloEvent:
			// Ignored
		case *slack.ConnectedEvent:
			log.WithFields(log.Fields{"Connection Counter:": ev.ConnectionCount, "Infos": ev.Info})
		case *slack.MessageEvent:
			log.WithFields(log.Fields{"Channel": ev.Channel, "message": ev.Text}).Debug("message event:")
			// send message to parser func
			out, err := parse(ev)
			if err != nil {
				log.WithField("ERROR", err).Error("parse message failed")
			}
			switch {
			case out.isEphemeral == true:
				params := slack.PostMessageParameters{
					AsUser: true,
				}
				_, err := postEphemeral(rtm, out.channel, out.user, out.message, params)
				if err != nil {
					log.WithField("ERROR", err).Error("Issue posting ephemeral message")
				}
			case out.isEphemeral == false:
				rtm.SendMessage(rtm.NewOutgoingMessage(out.message, out.channel))
			}
		case *slack.LatencyReport:
			log.WithField("Latency", ev.Value).Debug("Latency Reported")
		case *slack.RTMError:
			log.WithField("ERROR", ev.Error()).Error("RTM Error")
		case *slack.InvalidAuthEvent:
			log.Error("Invalid Credentials")
			return
		default:
			log.WithField("Data", ev).Debug("Some other data type")

		}

	}

}

// Cleans up Ephemeral message posting, see issue: https://github.com/nlopes/slack/issues/191
func postEphemeral(rtm *slack.RTM, channel, user, text string, params slack.PostMessageParameters) (string, error) {
	return rtm.PostEphemeral(
		channel,
		user,
		slack.MsgOptionText(text, params.EscapeText),
		slack.MsgOptionAttachments(params.Attachments...),
		slack.MsgOptionPostMessageParameters(params),
	)
}
