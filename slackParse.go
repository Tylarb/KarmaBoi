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
	"time"

	"github.com/nlopes/slack"
	log "github.com/sirupsen/logrus"
	"github.com/tylarb/TimeCache"
)

// set URL for expansion here
// TODO: change to OS env variable and/or move to SDFC api
const baseURL = "http://example.com/"

// Timeout in seconds to prevent karma spam
const timeout = 5 * 60

// Global rankings switch
const (
	TOP = iota
	BOTTOM
	SHAME
)

type response struct {
	message     string
	user        string
	channel     string
	isEphemeral bool
	isIM        bool
	threadTS    string
}

type karmaVal struct {
	name    string
	points  int
	shame   bool
	present bool // is the name present in the database
}

type cacheKey struct {
	*string
}

var cache = timeCache.NewSliceCache(timeout)

// regex definitions

var (
	karmaUp            = regexp.MustCompile(`.+\+{2}$`)      // name++
	karmaDown          = regexp.MustCompile(`.+-{2}$`)       // name--
	shameUp            = regexp.MustCompile(`.+~{2}$`)       // name~~
	nonKarmaWord       = regexp.MustCompile(`^\W+$`)         // not word characters (== [^0-9A-Za-z_])
	nonKarmaSingle     = regexp.MustCompile(`^.{1}$`)        // single character (so we exclude, for example, c++)
	nonKarmaPermission = regexp.MustCompile(`^[-d][-rwx]+$`) // "permissions" such as -rwxr-xr--
	question           = regexp.MustCompile(`.+\?{1,1}$`)    // keyword?
	weblink            = regexp.MustCompile(`^<http.+>$`)    // slack doesn't handle printing <link>

)

// 5 number long, beginning with 7 or 8
var caseID = regexp.MustCompile(`^[7-8][0-9]{4,4}$`)

// checks to see if the word being read is anything which we should ignore giving karma
func validKarmaCheck(s string) bool {
	valid := true
	switch {
	case s == "":
		valid = false
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
	var atBot = usrFormat(botID)

	if ev.User == "USLACKBOT" || ev.SubType == "bot_message" {
		log.Debug("Bot sent a message which is ignored")
		return nil
	}
	if strings.Contains(ev.Text, "beer") || strings.Contains(ev.Text, "I need a drink") {
		resp := slack.ItemRef{Channel: ev.Channel, Timestamp: ev.Timestamp}
		sc.AddReaction("beers", resp)
	}
	if strings.Contains(ev.Text, "wine") {
		resp := slack.ItemRef{Channel: ev.Channel, Timestamp: ev.Timestamp}
		sc.AddReaction("wine_glass", resp)
	}
	words := strings.Split(ev.Text, " ")
	switch {
	case words[0] == atBot:
		log.WithField("Message", ev.Text).Debug("Instuction for bot")
		err = handleCommand(ev, words)
	case len(words) == 1 && question.MatchString(words[0]):
		word := strings.Trim(words[0], "?")
		if also := isAlsoAsk(word); also != "" {
			if weblink.MatchString(also) {
				also = strings.Trim(also, "<>")
				also = strings.Split(also, "|")[0]
			}
			message := fmt.Sprintf("I recall hearing that %s is also %s", word, also)
			r := response{message: message, channel: ev.Channel}
			if ev.Timestamp != ev.ThreadTimestamp {
				r.threadTS = ev.ThreadTimestamp
			}
			slackPrint(r)
		}
	default:
		err = handleWord(ev, words)
	}
	return nil
}

// regex match and take appropriate action on words in a sentance. This only gets executed if
// the message is not deemed some other "type" of interation - like a command to the bot
func handleWord(ev *slack.MessageEvent, words []string) (err error) {

	var (
		s          string
		count      int
		message    string
		k          *karmaVal
		key        string    // key = user + target to prevent vote spam
		tc         bool      // time key was added to the cache
		r          time.Time // time remaining until able to be upvoted
		retMessage response
	)
	retMessage.channel = ev.Channel
	if ev.Timestamp != ev.ThreadTimestamp {
		retMessage.threadTS = ev.ThreadTimestamp
	}
	retArray := []string{}
	caseLinks := []string{}

	for _, word := range words {
		switch {

		case karmaUp.MatchString(word):
			k = newKarma(strings.Trim(word, "+"), false)
			if !validKarmaCheck(k.name) {
				continue
			}
			if k.name == usrFormat(ev.User) {
				k.shame = true
				k.modify(UP)
				s = fmt.Sprintf("Self promotion will get you nowhere.\n %s now has %d points of shame forever\n", k.name, k.points)
				retArray = append(retArray, s)
				continue
			}
			key = keygen(ev.User, k.name)
			tc, r = cache.Contains(key)
			if tc {
				timeWarn(ev, k.name, r)
			} else {
				k.modify(UP)
				s = responseGen(k, 0)
				retArray = append(retArray, s)
				getPrize(ev, k)
			}
			count++

		case karmaDown.MatchString(word):
			k = newKarma(strings.Trim(word, "-"), false)
			if !validKarmaCheck(k.name) {
				continue
			}
			if k.name == usrFormat(ev.User) {
				retArray = append(retArray, "Just remember that I will always love you and think you deserve all the karma")
			}
			key = keygen(ev.User, k.name)
			tc, r = cache.Contains(key)
			if tc {
				timeWarn(ev, k.name, r)
			} else {
				k.modify(DOWN)
				s = responseGen(k, 0)
				retArray = append(retArray, s)
				getPrize(ev, k)
			}
			count++

		case shameUp.MatchString(word):
			k = newKarma(strings.Trim(word, "~"), true)
			if !validKarmaCheck(k.name) {
				continue
			}
			if k.name == usrFormat(ev.User) {
				retArray = append(retArray, "I don't know why you're doing this to yourself but you probably deserve it")
			}
			key = keygen(ev.User, k.name)
			tc, r = cache.Contains(key)
			if tc {
				timeWarn(ev, k.name, r)
			} else {
				k.modify(UP)
				s = responseGen(k, 0)
				retArray = append(retArray, s)
			}
			count++

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
	retMessage.message = message
	err = slackPrint(retMessage)
	if err != nil {
		log.WithField("Err", err).Error("unable to print message to slack")
		return err
	}
	return nil

}

// Commands directed at the bot
func handleCommand(ev *slack.MessageEvent, words []string) error {
	retArray := []string{}
	var s string
	var err error
	var k *karmaVal
	var rank int
	r := response{channel: ev.Channel}

	switch {
	case len(words) > 2 && words[1] == "rank": // individual karma rankings
		for i := 2; i < len(words); i++ {
			k = newKarma(words[i], false)
			k.ask()
			rank = k.rank()
			if rank == 0 {
				continue
			}
			getPrize(ev, k)
			s = responseGen(k, rank)
			retArray = append(retArray, s)
		}
		r.message = strings.Join(retArray[:], "")

	case len(words) > 2 && words[1] == "rank~": // individual shame ranking
		for i := 2; i < len(words); i++ {
			k = newKarma(words[i], true)
			k.ask()
			rank = k.rank()
			if rank == 0 {
				continue
			}
			s = responseGen(k, rank)
			retArray = append(retArray, s)
		}
		r.message = strings.Join(retArray[:], "")
	case len(words) == 2 && words[1] == "rank":
		rankings := globalRank(TOP)
		r.message = rankingsPrint(rankings, TOP)

	case len(words) == 2 && words[1] == "!rank":
		rankings := globalRank(BOTTOM)
		r.message = rankingsPrint(rankings, BOTTOM)

	case len(words) == 2 && words[1] == "~rank":
		rankings := globalRank(SHAME)
		r.message = rankingsPrint(rankings, SHAME)
	case len(words) > 2 && words[1] == "list":
		if words[2] == "emails" {
			emails := getChanEmails(ev)
			r.message = fmt.Sprintf("```%s```", strings.Join(emails, ", "))
			r.threadTS = ev.Timestamp
		}
	case len(words) > 4 && words[2] == "is" && words[3] == "also":
		r.message = "I'll keep that in mind"
		also := strings.Join(words[4:], " ")
		isAlsoAdd(words[1], also)
	}

	err = slackPrint(r)
	if err != nil {
		log.WithField("Err", err).Error("unable to print message to slack")
	}

	return nil
}

func timeWarn(ev *slack.MessageEvent, n string, t time.Time) {
	tRemain := time.Duration(timeout)*time.Second - time.Since(t)
	message := fmt.Sprintf("Please wait %v before adjusting the karma of %s", tRemain, n)
	var r = response{message: message, user: ev.User, channel: ev.Channel, isEphemeral: true}
	slackPrint(r)
}

func newKarma(name string, shame bool) *karmaVal {
	k := new(karmaVal)
	k.name = name
	k.shame = shame
	return k
}

func keygen(u string, t string) string {
	s := []string{u, t}
	k := strings.Join(s, "-")
	return k
}

func responseGen(k *karmaVal, rank int) (s string) {
	if rank == 0 {
		switch {
		case k.shame && k.points == 1:
			s = fmt.Sprintf("What is done cannot be undone. %s now has shame forever\n", k.name)
		case k.shame:
			s = fmt.Sprintf("%s now has %d points of shame\n", k.name, k.points)
		default:
			s = fmt.Sprintf("%s now has %d points of karma\n", k.name, k.points)
		}
	} else {
		if k.shame {
			s = fmt.Sprintf("%s is rank %d with %d points of shame\n", k.name, rank, k.points)
		} else {
			s = fmt.Sprintf("%s is rank %d with %d points of karma\n", k.name, rank, k.points)
		}
	}
	return
}

func rankingsPrint(rankings []karmaVal, kind int) string {
	var (
		tank = ":fiestaparrot: :fiestaparrot: :fiestaparrot: TOP KARMA LEADERBOARD :fiestaparrot: :fiestaparrot: :fiestaparrot:\n"
		bank = ":sadparrot: :sadparrot: :sadparrot: BOTTOM KARMA LEADERBOARD :sadparrot: :sadparrot: :sadparrot:\n"
		sank = ":darth: :darth: :darth: SHAME LEADERBOARD darth: :darth: :darth:\n"
	)
	ranked := make([]string, 6) // Add space for header string at ranked[0]
	for i, k := range rankings {
		ranked[i+1] = fmt.Sprintf("%d. %s with %d\n", i+1, k.name, k.points)
	}
	switch {
	case kind == TOP:
		ranked[0] = tank
	case kind == BOTTOM:
		ranked[0] = bank
	case kind == SHAME:
		ranked[0] = sank
	}
	out := strings.Join(ranked, "")
	return out
}
