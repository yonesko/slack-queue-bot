package app

import (
	"fmt"
	"testing"
	"time"
)

func Test_humanizeDuration(t *testing.T) {
	tests := []struct {
		duration time.Duration
		want     string
	}{
		{0, ""},
		{time.Second, "1s"},
		{time.Minute, "1m"},
		{time.Minute + time.Second, "1m1s"},
		{time.Minute + 1, "1m"},
		{time.Hour * 77, "3d5h"},
		{time.Hour*77 + time.Minute*59, "3d5h59m"},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprint(tt), func(t *testing.T) {
			if got := humanizeDuration(tt.duration); got != tt.want {
				t.Errorf("humanizeDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}
