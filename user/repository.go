package user

import (
	"github.com/nlopes/slack"
	"github.com/yonesko/slack-queue-bot/model"
)

type Repository interface {
	FindById(id string) (model.User, error)
}

type repository struct {
	slackApi *slack.Client
}

func (r repository) FindById(id string) (model.User, error) {
	info, err := r.slackApi.GetUserInfo(id)
	if err != nil {
		return model.User{}, nil
	}
	return model.User{Id: id, FullName: info.RealName, DisplayName: info.Name}, nil
}

func NewRepository(slackApi *slack.Client) Repository {
	return repository{slackApi: slackApi}
}
