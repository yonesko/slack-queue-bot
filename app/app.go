package app

import (
	"fmt"
	"github.com/nlopes/slack"
	"github.com/yonesko/slack-queue-bot/queue"
	"github.com/yonesko/slack-queue-bot/usecase"
	"github.com/yonesko/slack-queue-bot/user"
	"gopkg.in/natefinch/lumberjack.v2"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
)

const thisBotUserId = "<@USMRFHHPE>"

type App struct {
	userRepository user.Repository
	queueService   usecase.QueueService
	lumberWriter   lumberjack.Logger
	rtm            *slack.RTM
	logger         *log.Logger
	controller     *Controller
}

func NewApp() *App {
	var lumberWriter = &lumberjack.Logger{
		Filename: "slack-queue-bot.log",
		MaxSize:  500,
		Compress: true,
	}
	log.SetOutput(lumberWriter)
	slackApi := slack.New(
		mustGetEnv("BOT_USER_OAUTH_ACCESS_TOKEN"),
		slack.OptionDebug(true),
		slack.OptionLog(log.New(lumberWriter, "slack_api: ", log.Lshortfile|log.LstdFlags)),
	)
	logger := log.New(lumberWriter, "queue-bot: ", log.Lshortfile|log.LstdFlags)
	userRepository := user.NewRepository(slackApi)
	return &App{
		userRepository: userRepository,
		lumberWriter:   lumberjack.Logger{},
		rtm:            connectToRTM(slackApi),
		logger:         logger,
		controller:     newController(userRepository, usecase.NewQueueService(queue.NewRepository())),
	}
}

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

func (app *App) Run() {
	app.printOnHello()
	for msg := range app.rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			if !needProcess(ev) {
				break
			}
			responseText := app.controller.execute(extractCommand(ev))
			app.rtm.SendMessage(app.rtm.NewOutgoingMessage(responseText, ev.Channel, slack.RTMsgOptionTS(ev.ThreadTimestamp)))
		case *slack.OutgoingErrorEvent:
			app.logger.Printf("Can't send msg: %s\n", ev.Error())

		}
	}
}

func (app *App) printOnHello() {
	bytes, err := ioutil.ReadFile("banner.txt")
	if err != nil {
		app.logger.Println(fmt.Errorf("can't read banner: %s", err))
		return
	}
	app.logger.Println(string(bytes))
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
	}
	return usecase.HelpCommand{}
}

func mustGetEnv(key string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	panic(fmt.Sprintf("environment variable %s unset", key))
}
