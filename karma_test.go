package main

import (
	"testing"

	"github.com/nlopes/slack"
)

// test 1: upvote, downvote, shame test
var TestMsgKarma1 = new(slack.Msg)

var TestMessageEvent1 = new(slack.MessageEvent)

// TestKarma checks adding and subtracting karma and shame, as well as ignored message and a case
func TestKarma(t *testing.T) {
	TestMsgKarma1.Channel = "testChannel"
	TestMsgKarma1.User = "testUser"
	TestMsgKarma1.Text = "test_up++ test_down-- test_shame~~ ###-- 71111"

	TestMessageEvent1.Msg = (*TestMsgKarma1)
	testRet, err := parse(TestMessageEvent1)
	if err != nil {
		t.Fatalf("Message event 1 (karma up/down test) failed with %v", err)
	}
	t.Logf("Return test 1 (karma up test): %v", testRet)

}
