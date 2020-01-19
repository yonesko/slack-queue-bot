package main

import (
	"fmt"
	"github.com/nlopes/slack/slacktest"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestHelp(t *testing.T) {
	s := slacktest.NewTestServer()
	go s.Start()
	s.SendMessageToBot("C123456789", "some text")
	expectedMsg := fmt.Sprintf("<@%s> %s", s.BotID, "some text")
	time.Sleep(2 * time.Second)
	assert.True(t, s.SawOutgoingMessage(expectedMsg))
	s.Stop()
}
