package model

type Queue struct {
	Users []User `json:"users"`
}

func (q Queue) IndexOf(user User) int {
	for i, u := range q.Users {
		if u.Id == user.Id {
			return i
		}
	}

	return -1
}

type User struct {
	Id string `json:"id"`
}
