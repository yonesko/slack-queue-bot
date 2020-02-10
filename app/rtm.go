package app

import (
	"fmt"
	"github.com/nlopes/slack"
	"github.com/yonesko/slack-queue-bot/usecase"
	"log"
	"strings"
	"time"
)

func connectToRTM(slackApi *slack.Client) *slack.RTM {
	rtm := slackApi.NewRTM()
	go rtm.ManageConnection()
	timeout := time.After(time.Second * 2)
	for {
		select {
		case <-timeout:
			log.Fatalf("timeouted while connect to Slack RTM")
		case msg := <-rtm.IncomingEvents:
			switch msg.Data.(type) {
			case *slack.InvalidAuthEvent, *slack.ConnectionErrorEvent:
				log.Fatal(fmt.Errorf("connection err: %s", msg))
			case *slack.HelloEvent:
				return rtm
			}
		}
	}
}

func needProcess(m *slack.MessageEvent) bool {
	mention := strings.HasPrefix(m.Text, thisBotUserId)
	isDirect := strings.HasPrefix(m.Channel, "D")
	simple := m.SubType == "" && !m.Hidden && m.BotID == "" && m.Edited == nil && m.User != ""
	return simple && (isDirect || mention)
}

func extractCommandTxt(text string) string {
	txt := strings.Replace(text, thisBotUserId, "", 1)
	txt = strings.ToLower(txt)
	return strings.TrimSpace(txt)
}

func extractCommand(ev *slack.MessageEvent) usecase.Command {
	return usecase.Command{AuthorUserId: ev.User, Data: extractData(ev)}
}

func extractData(ev *slack.MessageEvent) interface{} {
	switch extractCommandTxt(ev.Text) {
	case "add", "эд":
		return usecase.AddCommand{ToAddUserId: ev.User}
	case "del", "дел":
		return usecase.DelCommand{ToDelUserId: ev.User}
	case "show", "покаж":
		return usecase.ShowCommand{}
	case "clean":
		return usecase.CleanCommand{}
	case "pop":
		return usecase.PopCommand{}
	case "ack", "ак":
		return usecase.AckCommand{}
	case "pass", "пас":
		return usecase.PassCommand{}
	}
	return usecase.HelpCommand{}
}
