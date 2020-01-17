package main

import "testing"

func Test_extractCommand(t *testing.T) {
	tests := []struct {
		text string
		want string
	}{
		{text: "", want: ""},
		{text: "uhvbknjlm", want: "uhvbknjlm"},
		{text: "<@USMRFHHPE> add", want: "add"},
		{text: "<SOMEID> add", want: "add"},
		{text: "<@USMRFHHPE> someCmd \t", want: "someCmd"},
		{text: " someCmd", want: "someCmd"},
		{text: "add", want: "add"},
		{text: "5434424244", want: "5434424244"},
	}
	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			if got := extractCommand(tt.text); got != tt.want {
				t.Errorf("extractCommand() = %v, want %v", got, tt.want)
			}
		})
	}
}
