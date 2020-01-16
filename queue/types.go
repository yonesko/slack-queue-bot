package queue

type Queue struct {
	Users []User `json:"users"`
}

type User struct {
	Id      string `json:"id"`
	Channel string `json:"channel"`
}

func (q Queue) indexOf(user User) int {
	for i, u := range q.Users {
		if u == user {
			return i
		}
	}

	return -1
}
