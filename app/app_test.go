package app

import "testing"

/*
UCPHETPTJ
UPJ9Y95BM
*/
func Test_extractCommand(t *testing.T) {
	tests := []struct {
		text string
		want string
	}{
		{text: "", want: ""},
		{text: "uhvbknjlm", want: "uhvbknjlm"},
		{text: "<@USMRFHHPE> add", want: "add"},
		{text: "<@USMRFHHPE> someCmd \t", want: "somecmd"},
		{text: " someCmd", want: "somecmd"},
		{text: "add", want: "add"},
		{text: "5434424244", want: "5434424244"},
	}
	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			if got := extractCommandTxt(tt.text); got != tt.want {
				t.Errorf("extractCommandTxt() = %v, want %v", got, tt.want)
			}
		})
	}
}
