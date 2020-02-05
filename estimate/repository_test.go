package estimate

import (
	"bou.ke/monkey"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestEstimate_addOne(t *testing.T) {
	holdTime := time.Minute * 45
	estimate := Estimate{
		Average:     holdTime,
		Estimations: 1,
	}
	for i := 0; i < 100; i++ {
		estimate = estimate.AddOne(holdTime)
	}
	assert.Equal(t, 101, estimate.Estimations)
	assert.Equal(t, holdTime, estimate.Average)
}

func TestEstimate_YourHoldTsDontDependsOnNow(t *testing.T) {
	now := time.Now()
	patch := monkey.Patch(time.Now, func() time.Time { return now })
	defer patch.Unpatch()

	e := Estimate{Average: time.Minute * 45}
	holdTs := now
	yourHoldTs := now.Add(e.Average * 5)
	assert.Equal(t, yourHoldTs, now.Add(e.TimeToWait(5, holdTs)))
	now = now.Add(time.Minute)
	assert.Equal(t, yourHoldTs, now.Add(e.TimeToWait(5, holdTs)))
	now = now.Add(time.Minute * 30)
	assert.Equal(t, yourHoldTs, now.Add(e.TimeToWait(5, holdTs)))
}

func TestEstimate_TimeToWait(t *testing.T) {
	now := time.Now()
	patch := monkey.Patch(time.Now, func() time.Time { return now })
	defer patch.Unpatch()
	e := Estimate{Average: time.Minute * 45}
	type args struct {
		before    uint
		holdStart time.Time
	}

	tests := []struct {
		args args
		want time.Duration
	}{
		{args{before: 0, holdStart: now.Add(-time.Minute * 15)}, time.Duration(0)},
		{args{before: 1, holdStart: now.Add(-time.Minute * 15)}, time.Minute * 30},
		{args{before: 1, holdStart: now.Add(-time.Hour * 2)}, 0},
		{args{before: 2, holdStart: now.Add(-time.Minute * 15)}, time.Minute*30 + e.Average},
		{args{before: 2, holdStart: now.Add(-e.Average)}, e.Average},
		{args{before: 10, holdStart: now}, e.Average * 10},
	}
	for _, tt := range tests {
		name := fmt.Sprint(fmt.Sprintf("before %d %s", tt.args.before, now.Sub(tt.args.holdStart)))
		t.Run(name, func(t *testing.T) {
			if got := e.TimeToWait(tt.args.before, tt.args.holdStart); got != tt.want {
				t.Errorf("TimeToWait() = %v, want %v", got, tt.want)
			}
		})
	}
}
