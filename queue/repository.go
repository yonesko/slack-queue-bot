package queue

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type Repository interface {
	Save(Queue)
	Read() Queue
}

type fileRepository struct {
	filename string
}

func (f fileRepository) Save(queue Queue) {
	bytes, err := json.Marshal(queue)
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile(f.filename, bytes, 0644)
	if err != nil {
		panic(err)
	}
}

func (f fileRepository) Read() Queue {
	bytes, err := ioutil.ReadFile(f.filename)
	if os.IsNotExist(err) {
		return Queue{}
	}
	if err != nil {
		panic(err)
	}
	queue := &Queue{}
	err = json.Unmarshal(bytes, queue)
	if err != nil {
		panic(err)
	}
	return *queue
}
