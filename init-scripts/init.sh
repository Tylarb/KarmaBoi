#!/bin/bash 

# Released under MIT license, copyright 2018 Tyler Ramer

DIR=`pwd`
export BOT_HOME=~/.KarmaBoi

token=$BOT_HOME/bot_token


if [ -f $token ]; then
	echo "found bot token"
	export SLACK_BOT_NAME=$(grep name $token| cut -d : -f 2)
	export SLACK_BOT_TOKEN=$(grep token $token| cut -d : -f 2)
else
	echo "Please add your bot name and token at ~/.KarmaBoi/bot_token\nhttps://api.slack.com/bot-users\n\nFORMAT: \nname:[bot-name]\ntoken:[bot-token]"
fi

