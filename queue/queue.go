package queue

type User struct {
	Id string
}

type Queue struct {
	users []User
}

func (q Queue) indexOf(user User) int {
	for i, u := range q.users {
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

func (f fileRepository) Save(Queue) {
	panic("implement me")
}

func (f fileRepository) Read() Queue {
	panic("implement me")
}

type service struct {
	Repository
}

func (s service) Add(user User) {
	queue := s.Repository.Read()

	i := queue.indexOf(user)
	if i == -1 {
		queue.users = append(queue.users, user)
		s.Repository.Save(queue)
	}
}

func (s service) Delete(user User) {
	queue := s.Repository.Read()

	i := queue.indexOf(user)
	if i != -1 {
		queue.users = append(queue.users[:i], queue.users[:i+1]...)
		s.Repository.Save(queue)
	}
}

func (s service) Show() Queue {
	return s.Read()
}

func NewService() Service {
	return service{fileRepository{filename: "slack-queue-bot.db.json"}}
}
