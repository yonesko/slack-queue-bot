package main

import (
	_ "github.com/motemen/go-loghttp/global" //log HTTP req and resp
	"github.com/yonesko/slack-queue-bot/app"
	"github.com/yonesko/slack-queue-bot/i18n"
)

func main() {
	i18n.Init()
	app.NewApp().Run()
}
