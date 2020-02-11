package model

import "time"

type Queue struct {
	Entities         []QueueEntity `json:"entities"`
	HoldTs           time.Time     `json:"hold_ts"`
	HolderIsSleeping bool          `json:"holder_is_sleeping"`
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

func (q Queue) Index() map[string]int {
	ans := map[string]int{}
	for i, e := range q.Entities {
		ans[e.UserId] = i
	}
	return ans
}

func (q Queue) Copy() Queue {
	queue := Queue{Entities: make([]QueueEntity, len(q.Entities))}
	copy(queue.Entities, q.Entities)
	return queue
}

func (q Queue) CurHolder() string {
	if len(q.Entities) == 0 {
		return ""
	}
	return q.Entities[0].UserId
}
