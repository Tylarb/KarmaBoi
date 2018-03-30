/*
Text parsing for


Released under MIT license, copyright 2018 Tyler Ramer

*/

package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/nlopes/slack"
	log "github.com/sirupsen/logrus"
	"github.com/tylarb/TimeCache"
)

// set URL for expansion here
// TODO: change to OS env variable
const baseURL = "http://example.com/"
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
var karmaUp = regexp.MustCompile(`.+\+{2,2}$`)

// name--
var karmaDown = regexp.MustCompile(`.+-{2,2}$`)

// name~~
var shameUp = regexp.MustCompile(`.+~{2,2}$`)

// not word characters (== [^0-9A-Za-z_])
var nonKarma = regexp.MustCompile(`^\W+$`)

// 5 number long, beginning with 7 or 8
var caseID = regexp.MustCompile(`^[7-8][0-9]{4,4}$`)

// parses all messagess from slack for special commands or karma events
func parse(ev *slack.MessageEvent) (err error) {
	var atBot = fmt.Sprintf("<@%s>", botID)

	words := strings.Split(ev.Text, " ")
	switch {
	case words[0] == atBot:
		log.WithField("Message", ev.Text).Debug("Instuction for bot")
	default:
		err = handleWord(ev, words)
	}

	return nil
}

func handleWord(ev *slack.MessageEvent, words []string) (err error) {

	var s string
	var count int
	var k karmaVal
	var name string
	var message string

	retArray := []string{}
	caseLinks := []string{}

	for _, word := range words {
		switch {
		case nonKarma.MatchString(word):
			continue

		case karmaUp.MatchString(word):
			k.name = strings.Trim(word, "+")
			if nonKarma.MatchString(k.name) {
				continue
			}
			k.points, err = karmaAdd(name)
			s = responseGen(k)
			count++
			retArray = append(retArray, s)

		case karmaDown.MatchString(word):
			k.name = strings.Trim(word, "-")
			if nonKarma.MatchString(k.name) {
				continue
			}
			k.points, err = karmaSub(name)
			s = responseGen(k)
			count++
			retArray = append(retArray, s)

		case shameUp.MatchString(word):
			k.name = strings.Trim(word, "~")
			if nonKarma.MatchString(k.name) {
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
		message = fmt.Sprintf("%s gave various karma\n", usrFormat(ev.User))
	} else {
		retArray = append(retArray, caseLinks[:]...)
		message = strings.Join(retArray[:], "")
	}
	var r = response{message, ev.User, ev.Channel, false, false}
	err = scPrint(r)
	if err != nil {
		log.Error("unable to print message to slack")
		return err
	}
	return nil

}

func scPrint(r response) (err error) {
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
