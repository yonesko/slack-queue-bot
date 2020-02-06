package usecase

import (
	"errors"
	"github.com/yonesko/slack-queue-bot/model"
)

type QueueService interface {
	Add(model.QueueEntity) error
	DeleteById(toDelUserId string, authorUserId string) error
	Pop(authorUserId string) (string, error)
	Ack(authorUserId string) error
	PassFromSleepingHolder(authorUserId string) error
	DeleteAll() error
	Show() (model.Queue, error)
	UpdateOnNewHolder() error
}

var (
	AlreadyExistErr     = errors.New("already exist")
	NoSuchUserErr       = errors.New("no such user")
	QueueIsEmpty        = errors.New("queue is empty")
	HolderIsNotSleeping = errors.New("holder is not sleeping")
	YouAreNotHolder     = errors.New("you are not holder")
	NoOneToPass         = errors.New("no one to pass")
)
