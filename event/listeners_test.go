package event

import (
	"github.com/stretchr/testify/assert"
	"github.com/yonesko/slack-queue-bot/estimate"
	"strconv"
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
	assert.Equal(t, estimate.Estimate{time.Second * 100, 1}, duration)
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
	assert.Equal(t, estimate.Estimate{time.Second * 100, 1}, duration)
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
	assert.Equal(t, estimate.Estimate{time.Second * 0, 0}, duration)
}

func TestHoldTimeEstimateListener4(t *testing.T) {
	rep := &estimate.RepositoryMock{}
	listener := NewHoldTimeEstimateListener(rep)

	for i := 1; i <= 100; i++ {
		listener.Fire(NewHolderEvent{
			CurrentHolderUserId: strconv.Itoa(i),
			PrevHolderUserId:    strconv.Itoa(i - 1),
			AuthorUserId:        strconv.Itoa(i - 1),
			ts:                  time.Unix(int64(i)*77, 0),
		})
	}

	duration, err := rep.Get()
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, estimate.Estimate{time.Second * 77, 99}, duration)
}