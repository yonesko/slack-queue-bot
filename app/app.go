package app

import (
	"fmt"
	"github.com/nlopes/slack"
	"github.com/yonesko/slack-queue-bot/estimate"
	"github.com/yonesko/slack-queue-bot/event"
	"github.com/yonesko/slack-queue-bot/queue"
	"github.com/yonesko/slack-queue-bot/usecase"
	"github.com/yonesko/slack-queue-bot/user"
	"gopkg.in/natefinch/lumberjack.v2"
	"io/ioutil"
	"log"
	"os"
)

const (
	thisBotUserId = "<@USMRFHHPE>" //test bot user USG0TPHGA
	version       = "1.3.2"
)

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
	userRepository := user.NewRepository(slackApi)
	newHolderEventListeners := []event.NewHolderEventListener{
		event.NewNotifyNewHolderEventListener(slackApi, userRepository),
		event.NewHoldTimeEstimateListener(estimate.NewRepository()),
	}
	return &App{
		userRepository: userRepository,
		lumberWriter:   lumberjack.Logger{},
		rtm:            connectToRTM(slackApi),
		logger:         log.New(lumberWriter, "app: ", log.Lshortfile|log.LstdFlags),
		controller: newController(
			lumberWriter,
			userRepository,
			usecase.NewQueueService(
				queue.NewRepository(),
				event.NewQueueChangedEventBus(lumberWriter, newHolderEventListeners),
			),
		),
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
	app.logger.Printf("version %s", version)
}

func mustGetEnv(key string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	panic(fmt.Sprintf("environment variable %s unset", key))
}
