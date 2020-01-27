package queue

import (
	"encoding/json"
	"github.com/yonesko/slack-queue-bot/model"
	"io/ioutil"
	"os"
)

type Repository interface {
	Save(model.Queue) error
	Read() (model.Queue, error)
}

type fileRepository struct {
	filename string
}

func NewRepository() *fileRepository {
	createDbIfNeed()
	return &fileRepository{filename: "db/slack-queue-bot.db.json"}
}
func createDbIfNeed() {
	if _, err := os.Stat("db"); os.IsNotExist(err) {
		err := os.Mkdir("db", os.ModePerm)
		if err != nil {
			panic(err)
		}
	}
}
func (f *fileRepository) Save(queue model.Queue) error {
	bytes, err := json.Marshal(queue)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(f.filename, bytes, 0644)
	if err != nil {
		return err
	}
	return nil
}

func (f *fileRepository) Read() (model.Queue, error) {
	bytes, err := ioutil.ReadFile(f.filename)
	if os.IsNotExist(err) {
		return model.Queue{}, nil
	}
	if err != nil {
		return model.Queue{}, err
	}
	queue := &model.Queue{}
	err = json.Unmarshal(bytes, queue)
	if err != nil {
		return model.Queue{}, err
	}
	return *queue, nil
}
