package model

import (
	"github.com/magiconair/properties/assert"
	"testing"
)

func TestQueue_UserIdIndex(t *testing.T) {
	queue := Queue{}
	assert.Equal(t, map[string]int{}, queue.UserIdIndex())
	queue = Queue{Entities: []QueueEntity{{"1"}}}
	assert.Equal(t, map[string]int{
		"1": 0,
	}, queue.UserIdIndex())
	queue = Queue{Entities: []QueueEntity{{"1"}, {"2"}}}
	assert.Equal(t, map[string]int{
		"1": 0,
		"2": 1,
	}, queue.UserIdIndex())
}
