package queue

type Queue struct {
	Users []User `json:"users"`
}

type User struct {
	Id string `json:"id"`
}

func (q Queue) indexOf(user User) int {
	for i, u := range q.Users {
		if u.Id == user.Id {
			return i
		}
	}

	return -1
}
