package mock

import "github.com/yonesko/slack-queue-bot/model"

type UserRepository struct {
	data map[string]model.User
}

func NewUserRepository(data map[string]model.User) UserRepository {
	return UserRepository{data: data}
}

func (m UserRepository) FindById(id string) (model.User, error) {
	if user, ok := m.data[id]; ok {
		return user, nil
	}
	return model.User{}, nil
}
