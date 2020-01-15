package queue

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type User struct {
	Id string `json:"id"`
}

type Queue struct {
	Users []User `json:"users"`
}

func (q Queue) indexOf(user User) int {
	for i, u := range q.Users {
		if u == user {
			return i
		}
	}

	return -1
}

type Service interface {
	Add(User)
	Delete(User)
	Show() Queue
}

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

type service struct {
	Repository
}

func (s service) Add(user User) {
	queue := s.Repository.Read()

	i := queue.indexOf(user)
	if i == -1 {
		queue.Users = append(queue.Users, user)
		s.Repository.Save(queue)
	}
}

func (s service) Delete(user User) {
	queue := s.Repository.Read()
	i := queue.indexOf(user)
	if i != -1 {
		queue.Users = append(queue.Users[:i], queue.Users[i+1:]...)
		s.Repository.Save(queue)
	}
}

func (s service) Show() Queue {
	return s.Read()
}

func NewService() Service {
	return service{fileRepository{filename: "slack-queue-bot.db.json"}}
}
