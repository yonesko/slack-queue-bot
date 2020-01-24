package main

import (
	"fmt"
	_ "github.com/motemen/go-loghttp/global" //log HTTP req and resp
	"github.com/yonesko/slack-queue-bot/i18n"
	"github.com/yonesko/slack-queue-bot/queue"
	"github.com/yonesko/slack-queue-bot/usecase"
	"github.com/yonesko/slack-queue-bot/user"
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
	slackApi := slack.New(
		mustGetEnv("BOT_USER_OAUTH_ACCESS_TOKEN"),
		slack.OptionDebug(true),
		slack.OptionLog(log.New(lumberWriter, "slack_api: ", log.Lshortfile|log.LstdFlags)),
	)
	rtm := slackApi.NewRTM()
	go rtm.ManageConnection()
	userRepository := user.NewRepository(slackApi)
	queueRepository := queue.NewRepository()
	controller := newController(userRepository, queueRepository)
	logger := log.New(lumberWriter, "queue-bot: ", log.Lshortfile|log.LstdFlags)
	logger.Println("Service is started")
	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			if !needProcess(ev) {
				break
			}
			responseText, err := controller.execute(extractCommand2(ev))
			if err != nil {
				responseText = i18n.P.MustGetString("error_occurred")
				logger.Println(err)
			}
			rtm.SendMessage(rtm.NewOutgoingMessage(responseText, ev.Channel, slack.RTMsgOptionTS(ev.ThreadTimestamp)))
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

func extractCommand2(ev *slack.MessageEvent) usecase.Command {
	return usecase.Command{
		AuthorUserId: ev.User,
		Data:         extractData(ev),
	}
}

func extractData(ev *slack.MessageEvent) interface{} {
	switch extractCommand(ev.Text) {
	case "add":
		return usecase.AddCommand{ToAddUserId: ev.User}
	case "del":
		return usecase.DelCommand{ToDelUserId: ev.User}
	case "show":
		return usecase.ShowCommand{}
	case "clean":
		return usecase.CleanCommand{}
	case "pop":
		return usecase.PopCommand{}
	default:
		return usecase.HelpCommand{}
	}
}

func mustGetEnv(key string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	panic(fmt.Sprintf("environment variable %s unset", key))
}
