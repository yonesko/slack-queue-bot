package model

import (
	"github.com/magiconair/properties/assert"
	"testing"
)

func TestQueue_Index(t *testing.T) {
	queue := Queue{}
	assert.Equal(t, map[string]int{}, queue.Index())
	queue = Queue{Entities: []QueueEntity{{"1"}}}
	assert.Equal(t, map[string]int{
		"1": 0,
	}, queue.Index())
	queue = Queue{Entities: []QueueEntity{{"1"}, {"2"}}}
	assert.Equal(t, map[string]int{
		"1": 0,
		"2": 1,
	}, queue.Index())
}
