package queue

type Service interface {
	Add(User)
	Delete(User)
	Show() Queue
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
