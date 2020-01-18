package queue

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"sync"
)

type Repository interface {
	Save(Queue) error
	Read() (Queue, error)
}

type fileRepository struct {
	filename string
	queue    Queue
	mtx      sync.Mutex
}

func (f *fileRepository) Save(queue Queue) error {
	f.mtx.Lock()
	defer f.mtx.Unlock()
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
	f.mtx.Lock()
	defer f.mtx.Unlock()
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
