package model

type User struct {
	ID   int         `json:"id"`
	UUID string      `json:"uuid"`
	URL  []UserURLID `json:"url"`
}

type UserURLID struct {
	ID int `json:"id"`
}
