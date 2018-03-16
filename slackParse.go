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
)

// set URL for expansion here
const baseURL = "http://example.com/"

type retMessage struct {
	message     string
	isEphemeral bool
	user        string
	channel     string
}

type karmaVal struct {
	name   string
	points int
	shame  bool
}

// regex definitions

var karmaUp = regexp.MustCompile(`.+\+{2,2}$`)
var karmaDown = regexp.MustCompile(`.+-{2,2}$`)
var shameUp = regexp.MustCompile(`.+~{2,2}$`)
var nonKarma = regexp.MustCompile(`^\W+$`)
var caseID = regexp.MustCompile(`^[7-8][0-9]{4,4}$`)

// parses all messagess from slack for special commands or karma events
func parse(message *slack.MessageEvent) (retMessage, error) {
	var s string
	var count int
	var k karmaVal
	var name string
	var response string
	var err error
	words := strings.Split(message.Text, " ")
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
		response = fmt.Sprintf("%s gave various karma\n", slackFMT(message.User))
	} else {
		retArray = append(retArray, caseLinks[:]...)
		response = strings.Join(retArray[:], "")
	}
	out := retMessage{response, false, message.User, message.Channel}
	return out, nil
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

func slackFMT(u string) string {
	return fmt.Sprintf("<@%s>", u)
}
