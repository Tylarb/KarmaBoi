/*
Generic useful auxiliary functions
*/

package main

import (
	"fmt"

	"github.com/nlopes/slack"
	log "github.com/sirupsen/logrus"
)

// gets an array of email addresses for all users in channel
func getChanEmails(ev *slack.MessageEvent) []string {
	c := ev.Channel
	params := slack.GetUsersInConversationParameters{ChannelID: c}
	users, _, err := sc.GetUsersInConversation(&params)
	emails := []string{}
	if err != nil {
		log.Fatal(err)
	}
	for _, user := range users {
		info, err := sc.GetUserInfo(user)
		if err != nil {
			log.Fatal(err)
		}
		if email := info.Profile.Email; email != "" {
			emails = append(emails, email)
		}
	}
	return emails

}

func usrFormat(u string) string {
	return fmt.Sprintf("<@%s>", u)
}

// Print messages to slack. Accepts response struct and returns any errors on the print
func slackPrint(r response) (err error) {
	switch {
	case r.isEphemeral:
		_, err = postEphemeral(rtm, r.channel, r.user, r.message)
	default:
		rtm.SendMessage(rtm.NewOutgoingMessage(r.message, r.channel, slack.RTMsgOptionTS(r.threadTS)))
		err = nil
	}
	return
}
