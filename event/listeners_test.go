package event

import (
	"github.com/stretchr/testify/assert"
	"github.com/yonesko/slack-queue-bot/estimate"
	"testing"
	"time"
)

func TestHoldTimeEstimateListener1(t *testing.T) {
	rep := &estimate.RepositoryMock{}
	listener := NewHoldTimeEstimateListener(rep)

	listener.Fire(NewHolderEvent{
		CurrentHolderUserId: "123",
		PrevHolderUserId:    "",
		AuthorUserId:        "123",
		ts:                  time.Unix(0, 0),
	})

	listener.Fire(NewHolderEvent{
		CurrentHolderUserId: "abc",
		PrevHolderUserId:    "123",
		AuthorUserId:        "123",
		ts:                  time.Unix(100, 0),
	})
	duration, err := rep.Get()
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, duration, time.Second*100)
}

func TestHoldTimeEstimateListener2(t *testing.T) {
	rep := &estimate.RepositoryMock{}
	listener := NewHoldTimeEstimateListener(rep)

	listener.Fire(NewHolderEvent{
		CurrentHolderUserId: "1",
		PrevHolderUserId:    "2",
		AuthorUserId:        "1",
		ts:                  time.Unix(0, 0),
	})

	listener.Fire(NewHolderEvent{
		CurrentHolderUserId: "3",
		PrevHolderUserId:    "1",
		AuthorUserId:        "1",
		ts:                  time.Unix(100, 0),
	})
	duration, err := rep.Get()
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, duration, time.Second*100)
}
func TestHoldTimeEstimateListener3(t *testing.T) {
	rep := &estimate.RepositoryMock{}
	listener := NewHoldTimeEstimateListener(rep)

	listener.Fire(NewHolderEvent{
		CurrentHolderUserId: "1",
		PrevHolderUserId:    "2",
		AuthorUserId:        "1",
		ts:                  time.Unix(0, 0),
	})

	listener.Fire(NewHolderEvent{
		CurrentHolderUserId: "3",
		PrevHolderUserId:    "1",
		AuthorUserId:        "4",
		ts:                  time.Unix(100, 0),
	})
	duration, err := rep.Get()
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, duration, time.Second*0)
}
