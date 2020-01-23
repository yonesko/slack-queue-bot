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
