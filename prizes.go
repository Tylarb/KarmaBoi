/* ASCSII art */

package main

import (
	"math/rand"
	"strings"
	"time"

	"github.com/nlopes/slack"
)

const (
	ePrize = `
	    ___     _,.--.,_
     .-~   ~--"~-.   ._ "-.
    /      ./_    Y    "-. \
   Y       :~     !         Y
    l0 0    |     /          .|
 _   \. .-, l    /           |j
()\___) |/   \_/";          !
 \._____.-~\  .  ~\.      ./
		    Y_ Y_. "vr"~  T
		    (  (    |L    j
		    [nn[nn..][nn..]
	  ~~~~~~~~~~~~~~~~~~~~~~~
  ___  _ __   ___   _   _ _ __
 / _ \| '_ \ / _ \ | | | | '_ \
| (_) | | | |  __/ | |_| | |_) |
 \___/|_| |_|\___|  \__,_| .__/
					   | |
					   |_|
`

	rPrize = `
			  ,
			 /|      __
		    / |   ,-~ /
		   Y :|  //  /
		   | jj /( .^
		   >-"~"-v"
		  /       Y
		 jo  o    |
		( ~T~     j
		 >._-' _./
		/   "~"  |
		Y     _,  |
	   /| ;-"~ _  l
	  / l/ ,-"~    \
	  \//\/      .- \
	  Y        /    Y
	  l       I     !
	  ]\      _\    /"\\
	 (" ~----( ~   Y.  )
   ~~~~~~~~~~~~~~~~~~~~~~~~~
  ____  _      _____   _     ____
 /  _ \/ \  /|/  __/  / \ /\/  __\
|  / \|| |\ |||  \    | | |||  \/|
|  \_/|| | \|||  /_   | \_/||  __/
 \____/\_/  \|\____\  \____/\_/

`

	aPrize = `
________________________________
  __   __ _  ____    _  _  ____ \      ------
 /  \ (  ( \(  __)  / )( \(  _ \ \    | | # \                                 |
(  O )/    / ) _)   ) \/ ( ) __/   -- | ____ \________|----^"|"|"\___________ |
 \__/ \_)__)(____)  \____/(__)   /     \___\   FO + 94 >>    '""""    =====  "|D
________________________________/            ^^------____--""""""+""--_  __--"|
														'|"->##)+---|'""      |
																\  \
															   <- O -)
																 '"'
  
`
)

func getPrize(ev *slack.MessageEvent, k *karmaVal) {
	var prizes = []string{ePrize, aPrize, rPrize}
	var prize string
	if k.points > 0 {
		switch {
		case k.points%5000 == 0:

		case k.points%1000 == 0:

		case k.points%100 == 0:
			prize = prizes[rand.Intn(len(prizes))]
		}
	}
	go printPrize(ev, prize)
}

func printPrize(ev *slack.MessageEvent, prize string) {
	channel, timestamp, _ := sc.PostMessage(ev.Channel, slack.MsgOptionAsUser(true), slack.MsgOptionText("beep beep whirrrree", false))
	linebyline := strings.Split(prize, "\n")
	completedImage := []string{}
	for _, line := range linebyline {
		time.Sleep(200 * time.Millisecond)
		completedImage = append(completedImage, line)
		image := "```" + strings.Join(completedImage, "\n") + "```"
		sc.UpdateMessage(channel, timestamp, slack.MsgOptionText(image, false))
	}

}
