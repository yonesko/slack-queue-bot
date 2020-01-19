package queue

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type Repository interface {
	Save(Queue) error
	Read() (Queue, error)
}

type fileRepository struct {
	filename string
	queue    Queue
}

func newFileRepository() *fileRepository {
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
func (f *fileRepository) Save(queue Queue) error {
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

func (f *fileRepository) Read() (Queue, error) {
	bytes, err := ioutil.ReadFile(f.filename)
	if os.IsNotExist(err) {
		return Queue{}, nil
	}
	if err != nil {
		return Queue{}, err
	}
	queue := &Queue{}
	err = json.Unmarshal(bytes, queue)
	if err != nil {
		return Queue{}, err
	}
	return *queue, nil
}
