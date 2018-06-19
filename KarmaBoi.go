/*
Karma tracking bot with additional features. Integrates with Slack, runs on PCF


Released under MIT license, copyright 2018 Tyler Ramer

*/

package main

import (
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/cloudfoundry-community/go-cfenv"
	"github.com/nlopes/slack"
)

// CF database service
const serviceLable = "elephantsql"

var conStr string

// Get bot name and token from env, and make sure botID is globally accessible
var (
	slackBotToken = os.Getenv("SLACK_BOT_TOKEN")
	slackBotName  = os.Getenv("SLACK_BOT_NAME")
	botID         string
)

// The slack client and RTM messaging are used as an out - rather than passing
// the SC to each function, define it globally to ease accessed. We do handle
// errors in the main function, however

var (
	sc  *slack.Client
	rtm *slack.RTM
)

// log levels, see logrus docs
func init() {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)

	appEnv, err := cfenv.Current()
	if err != nil {
		log.Fatal("Could not get cloud foundry environment details")
	}
	dbService, err := appEnv.Services.WithLabel(serviceLable)
	if err != nil {
		log.Fatal(err)
	}

	if len(dbService) != 1 {
		log.WithField("serviceLable", serviceLable).Fatal("It appears that more than one service with the serviceLable is attached to your app\nPlease ensure there is only one database instance attached")
	}
	var ok bool
	conStr, ok = dbService[0].CredentialString("uri")
	if !ok {
		log.Fatal("Could not find database URI")
	}
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
	sc = slack.New(slackBotToken)
	botID = getBotID(slackBotName, sc)
	log.WithField("ID", botID).Debug("Bot ID returned")
	rtm = sc.NewRTM()
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
			err := parse(ev)
			if err != nil {
				log.WithField("ERROR", err).Error("parse message failed")
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
func postEphemeral(rtm *slack.RTM, channel, user, text string) (string, error) {
	params := slack.PostMessageParameters{
		AsUser: true,
	}
	return rtm.PostEphemeral(
		channel,
		user,
		slack.MsgOptionText(text, params.EscapeText),
		slack.MsgOptionAttachments(params.Attachments...),
		slack.MsgOptionPostMessageParameters(params),
	)
}
