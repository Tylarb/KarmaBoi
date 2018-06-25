/*
All RTM messages are sent here in order to be parsed. All commands are of one
of three types:

1. Message at bot > command function
2. Bookmark query of format <bookmark>? [one word long]
3. any other message, which is parsed for karma, or shame

Because all messages are either logs or printed to slack, a slack client is
defined at the main package in order to reduce passing the slack client around



Released under MIT license, copyright 2018 Tyler Ramer

*/

package main

import (
	"fmt"
	"math/rand"
	"regexp"
	"strings"

	"github.com/nlopes/slack"
	log "github.com/sirupsen/logrus"
	"github.com/tylarb/TimeCache"
)

// set URL for expansion here
// TODO: change to OS env variable and/or move to SDFC api
const baseURL = "http://example.com/"

// Timeout in seconds to prevent karma spam
const timeout = 5 * 60

type response struct {
	message     string
	user        string
	channel     string
	isEphemeral bool
	isIM        bool
}

type karmaVal struct {
	name   string
	points int
	shame  bool
}

var cache = timeCache.NewSliceCache(timeout)

// regex definitions

// name++
var karmaUp = regexp.MustCompile(`.+\+{2}$`)

// name--
var karmaDown = regexp.MustCompile(`.+-{2}$`)

// name~~
var shameUp = regexp.MustCompile(`.+~{2}$`)

// not word characters (== [^0-9A-Za-z_])
var nonKarmaWord = regexp.MustCompile(`^\W+$`)

// single character (so we exclude, for example, c++)
var nonKarmaSingle = regexp.MustCompile(`^.{1}$`)

// "permissions" such as -rwxr-xr--
var nonKarmaPermission = regexp.MustCompile(`^[-d][-rwx]+$`)

// 5 number long, beginning with 7 or 8
var caseID = regexp.MustCompile(`^[7-8][0-9]{4,4}$`)

// checks to see if the word being read is anything which we should ignore giving karma
func validKarmaCheck(s string) bool {
	valid := true
	switch {
	case nonKarmaWord.MatchString(s):
		valid = false
	case nonKarmaSingle.MatchString(s):
		valid = false
	case nonKarmaPermission.MatchString(s):
		valid = false
	}
	return valid
}

// parses all messagess from slack for special commands or karma events
func parse(ev *slack.MessageEvent) (err error) {
	var atBot = fmt.Sprintf("<@%s>", botID)
	if ev.User == "USLACKBOT" {
		log.Debug("Slackbot sent a message which is ignored")
		return nil
	}
	words := strings.Split(ev.Text, " ")
	switch {
	case words[0] == atBot:
		log.WithField("Message", ev.Text).Debug("Instuction for bot")
		err = handleCommand(ev, words)
	default:
		err = handleWord(ev, words)
	}

	return nil
}

// regex match and take appropriate action on words in a sentance. This only gets executed if
// the message is not deemed some other "type" of interation - like a command to the bot
func handleWord(ev *slack.MessageEvent, words []string) (err error) {

	var s string
	var count int
	var k karmaVal
	// TODO: Remove name here
	var name string
	var message string

	retArray := []string{}
	caseLinks := []string{}

	for _, word := range words {
		switch {

		case karmaUp.MatchString(word):
			k.name = strings.Trim(word, "+")
			if !validKarmaCheck(k.name) {
				continue
			}
			k.points, err = karmaAdd(name) // TODO: use karmaAdd (k karmaVal)
			s = responseGen(k)
			count++
			retArray = append(retArray, s)

		case karmaDown.MatchString(word):
			k.name = strings.Trim(word, "-")
			k.points, err = karmaSub(name)
			if !validKarmaCheck(k.name) {
				continue
			}
			s = responseGen(k)
			count++
			retArray = append(retArray, s)

		case shameUp.MatchString(word):
			k.name = strings.Trim(word, "~")
			if !validKarmaCheck(k.name) {
				continue
			}
			k.points, err = shameAdd(name)
			k.shame = true
			s = responseGen(k)
			count++
			retArray = append(retArray, s)

		case caseID.MatchString(word):
			caseLinks = append(caseLinks, baseURL, word, "\n")
		}
	}
	if err != nil {
		log.WithField("ERROR", err).Info("Failed to adjust karma")
	}
	if count > 3 {
		multiKarmaResponse := []string{
			"tosses karma like it's candy",
			"is feeling very generous with karma today",
			"is a karma distributing overlord",
		}
		n := rand.Int() % len(multiKarmaResponse)

		message = fmt.Sprintf("%s %s\n", usrFormat(ev.User), multiKarmaResponse[n])
	} else {
		retArray = append(retArray, caseLinks[:]...)
		message = strings.Join(retArray[:], "")
	}
	var r = response{message, ev.User, ev.Channel, false, false}
	err = slackPrint(r)
	if err != nil {
		log.WithField("Err", err).Error("unable to print message to slack")
		return err
	}
	return nil

}

// Commands directed at the bot
func handleCommand(ev *slack.MessageEvent, words []string) error {
	retArray := []string{}
	var k karmaVal
	var message string
	var s string
	var err error

	// individual rankings
	if len(words) > 2 && words[1] == "rank" {
		for i := 2; i < len(words); i++ {
			k = karmaVal{words[i], 0, false}
			k = karmaRank(k)
			s = responseGen(k)
			retArray = append(retArray, s)
		}
		message = strings.Join(retArray[:], "")
		var r = response{message, ev.User, ev.Channel, false, false}
		err = slackPrint(r)
		if err != nil {
			log.WithField("Err", err).Error("unable to print message to slack")
		}
	}

	return nil
}

// Print messages to slack. Accepts response struct and returns any errors on the print
func slackPrint(r response) (err error) {
	switch {
	case r.isEphemeral:
		_, err = postEphemeral(rtm, r.channel, r.user, r.message)
	default:
		rtm.SendMessage(rtm.NewOutgoingMessage(r.message, r.channel))
		err = nil
	}
	return
}

func karmaAdd(name string) (karma int, err error) {
	return 1, nil
}

func karmaSub(name string) (karma int, err error) {
	return 1, nil
}

func shameAdd(name string) (karma int, err error) {
	return 1, nil
}

func responseGen(k karmaVal) (s string) {
	if k.shame {
		if k.points == 1 {
			s = fmt.Sprintf("What is done cannot be undone. %s now has shame forever\n", k.name)
			return
		}
		s = fmt.Sprintf("%s now has %d points of shame\n", k.name, k.points)
		return
	}
	s = fmt.Sprintf("%s now has %d points of karma\n", k.name, k.points)
	return
}

func usrFormat(u string) string {
	return fmt.Sprintf("<@%s>", u)
}
