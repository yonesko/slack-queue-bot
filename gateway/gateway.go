package gateway

import (
	"github.com/nlopes/slack"
	"log"
)

type Gateway interface {
	Send(userId, txt string) error
	SendAndLog(userId, txt string)
}

type slackGateway struct {
	slackApi *slack.Client
}

func (s slackGateway) SendAndLog(userId, txt string) {
	err := s.Send(userId, txt)
	if err != nil {
		log.Printf("can't send %s '%s' %s", userId, txt, err)
	}
}

func NewSlackGateway(slackApi *slack.Client) *slackGateway {
	return &slackGateway{slackApi: slackApi}
}

func (s slackGateway) Send(userId, txt string) error {
	if userId == "" {
		log.Printf("sendMsg user id is empty")
		return nil
	}
	_, _, err := s.slackApi.PostMessage(userId,
		slack.MsgOptionText(txt, true),
		slack.MsgOptionAsUser(true),
	)
	return err
}
