package main

import (
	"fmt"
	"gopkg.in/natefinch/lumberjack.v2"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/nlopes/slack"
)

const thisBotUserId = "<@USMRFHHPE>"

var lumberWriter = &lumberjack.Logger{
	Filename: "slack-queue-bot.log",
	MaxSize:  500,
	Compress: true,
}

func main() {
	log.SetOutput(lumberWriter)
	logger := log.New(lumberWriter, "queue-bot: ", log.Lshortfile|log.LstdFlags)
	controller := newController()
	logger.Println("Service is started")
	for msg := range controller.rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			if !needProcess(ev) {
				break
			}
			controller.handleMessageEvent(ev)
		case *slack.OutgoingErrorEvent:
			logger.Printf("Can't send msg: %s\n", ev.Error())
		case *slack.InvalidAuthEvent, *slack.ConnectionErrorEvent:
			log.Fatal(fmt.Errorf("connection err: %s", msg))
		case *slack.HelloEvent:
			printOnHello(logger)
		}
	}
}

func printOnHello(logger *log.Logger) {
	logger.Println("Hello from Slack server received")
	bytes, err := ioutil.ReadFile("banner.txt")
	if err != nil {
		logger.Println(fmt.Errorf("can't read banner: %s", err))
		return
	}
	logger.Println(string(bytes))
}

func needProcess(m *slack.MessageEvent) bool {
	mention := strings.HasPrefix(m.Text, thisBotUserId)
	isDirect := strings.HasPrefix(m.Channel, "D")
	simple := m.SubType == "" && !m.Hidden && m.BotID == "" && m.Edited == nil && m.User != ""
	return simple && (isDirect || mention)
}

func extractCommand(text string) string {
	txt := strings.Replace(text, thisBotUserId, "", 1)
	txt = strings.ToLower(txt)
	return strings.TrimSpace(txt)
}

func mustGetEnv(key string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	panic(fmt.Sprintf("environment variable %s unset", key))
}
