package event

import (
	"github.com/stretchr/testify/assert"
	"github.com/yonesko/slack-queue-bot/estimate"
	"strconv"
	"testing"
	"time"
)

func TestHoldTimeEstimateListener_FirstInQueue(t *testing.T) {
	rep := &estimate.RepositoryMock{}
	listener := NewHoldTimeEstimateListener(rep)

	listener.Fire(NewHolderEvent{
		CurrentHolderUserId: "123",
		PrevHolderUserId:    "",
		AuthorUserId:        "123",
		Ts:                  time.Unix(0, 0),
	})

	listener.Fire(NewHolderEvent{
		CurrentHolderUserId: "abc",
		PrevHolderUserId:    "123",
		AuthorUserId:        "123",
		Ts:                  time.Unix(100, 0),
	})
	duration, err := rep.Get()
	assert.Nil(t, err)
	assert.Equal(t, estimate.Estimate{time.Second * 100, 1}, duration)
}

func TestHoldTimeEstimateListener_TooLongTime(t *testing.T) {
	rep := &estimate.RepositoryMock{}
	listener := NewHoldTimeEstimateListener(rep)

	listener.Fire(NewHolderEvent{
		CurrentHolderUserId: "123",
		PrevHolderUserId:    "",
		AuthorUserId:        "123",
		Ts:                  time.Unix(0, 0),
	})

	listener.Fire(NewHolderEvent{
		CurrentHolderUserId: "abc",
		PrevHolderUserId:    "123",
		AuthorUserId:        "123",
		Ts:                  time.Unix(int64((time.Hour * 2).Seconds()), 0),
	})
	duration, err := rep.Get()
	assert.Nil(t, err)
	assert.Equal(t, estimate.Estimate{0, 0}, duration)
}

func TestHoldTimeEstimateListener_InMiddleOfQueue(t *testing.T) {
	rep := &estimate.RepositoryMock{}
	listener := NewHoldTimeEstimateListener(rep)

	listener.Fire(NewHolderEvent{
		CurrentHolderUserId: "1",
		PrevHolderUserId:    "2",
		AuthorUserId:        "1",
		Ts:                  time.Unix(0, 0),
	})

	listener.Fire(NewHolderEvent{
		CurrentHolderUserId: "3",
		PrevHolderUserId:    "1",
		AuthorUserId:        "1",
		Ts:                  time.Unix(100, 0),
	})
	duration, err := rep.Get()
	assert.Nil(t, err)
	assert.Equal(t, estimate.Estimate{time.Second * 100, 1}, duration)
}

func TestHoldTimeEstimateListener_ForceDel(t *testing.T) {
	rep := &estimate.RepositoryMock{}
	listener := NewHoldTimeEstimateListener(rep)

	listener.Fire(NewHolderEvent{
		CurrentHolderUserId: "1",
		PrevHolderUserId:    "2",
		AuthorUserId:        "1",
		Ts:                  time.Unix(0, 0),
	})

	listener.Fire(NewHolderEvent{
		CurrentHolderUserId: "3",
		PrevHolderUserId:    "1",
		AuthorUserId:        "4",
		Ts:                  time.Unix(100, 0),
	})
	duration, err := rep.Get()
	assert.Nil(t, err)
	assert.Equal(t, estimate.Estimate{time.Second * 0, 0}, duration)
}

func TestHoldTimeEstimateListener_MultiplyEvents(t *testing.T) {
	rep := &estimate.RepositoryMock{}
	listener := NewHoldTimeEstimateListener(rep)

	for i := 1; i <= 100; i++ {
		listener.Fire(NewHolderEvent{
			CurrentHolderUserId: strconv.Itoa(i),
			PrevHolderUserId:    strconv.Itoa(i - 1),
			AuthorUserId:        strconv.Itoa(i - 1),
			Ts:                  time.Unix(int64(i)*77, 0),
		})
	}

	duration, err := rep.Get()
	assert.Nil(t, err)
	assert.Equal(t, estimate.Estimate{time.Second * 77, 99}, duration)
}
