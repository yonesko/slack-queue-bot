package main

import (
	_ "github.com/motemen/go-loghttp/global" //log HTTP req and resp
	"github.com/yonesko/slack-queue-bot/app"
)

func main() {
	app.NewApp().Run()
}
