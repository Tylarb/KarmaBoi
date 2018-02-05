package main

import (
	"fmt"
	"os"
)

var SLACK_BOT_TOKEN = os.Getenv("SLACK_BOT_TOKEN")
var SLACK_BOT_NAME = os.Getenv("SLACK_BOT_NAME")

func main() {
	fmt.Println("hello, world!")
	fmt.Println("slack token:", SLACK_BOT_TOKEN)
	fmt.Println("slack name:", SLACK_BOT_NAME)
}
