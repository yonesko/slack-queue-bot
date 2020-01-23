package model

type User struct {
	Id          string `json:"id"`
	FullName    string `json:"full_name"`
	DisplayName string `json:"display_name"`
}
