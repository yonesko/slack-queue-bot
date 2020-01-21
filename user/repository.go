package user

import "github.com/yonesko/slack-queue-bot/model"

type Repository interface {
	FindById(id string) (model.User, error)
}
