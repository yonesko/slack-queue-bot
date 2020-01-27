package estimate

import (
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
