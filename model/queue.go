package model

type Queue struct {
	Entities []QueueEntity `json:"entities"`
}

type QueueEntity struct {
	UserId string `json:"user_id"`
}

func (q Queue) IndexOf(userId string) int {
	for i, e := range q.Entities {
		if e.UserId == userId {
			return i
		}
	}

	return -1
}

func (q Queue) Copy() Queue {
	queue := Queue{Entities: make([]QueueEntity, len(q.Entities))}
	copy(queue.Entities, q.Entities)
	return queue
}
