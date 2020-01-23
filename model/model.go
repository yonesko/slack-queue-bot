package model

type Queue struct {
	Entities []QueueEntity `json:"entities"`
}

type QueueEntity struct {
	UserId string `json:"user_id"`
}

func (q Queue) IndexOf(ent QueueEntity) int {
	for i, e := range q.Entities {
		if e == ent {
			return i
		}
	}

	return -1
}

type User struct {
	Id          string `json:"id"`
	FullName    string `json:"full_name"`
	DisplayName string `json:"display_name"`
}
