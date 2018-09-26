package models

type (
	// User represents the structure of our resource
	User struct {
		Name string `json:"name"`
		Pass string `json:"pass"`
		Id   int64  `json:"id"`
	}
)

type (
	// User represents the structure of our resource
	UserNameHashFieldValue struct {
		Pass string `json:"pass"`
		Id   int64  `json:"id"`
	}
)
