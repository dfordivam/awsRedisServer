package models

type (
	// User
	User struct {
		Name string `json:"name"`
		Pass string `json:"pass"`
		Id   int64  `json:"id"`
	}
)

type (
	// User Data stored in Hash field value
	UserNameHashFieldValue struct {
		Pass string `json:"pass"`
		Id   int64  `json:"id"`
	}
)

type (
	// Message stored in DB
	MessageObject struct {
		User    string `json:"user"`
		Message string `json:"msg"`
	}
)

type (
	// Message stored in DB
	PostMessage struct {
		Message       string `json:"msg"`
		LastMessageId int64  `json:"id"`
	}
)

type (
	// Server Response
	SendMessages struct {
		Messages  []MessageObject `json:"msgs"`
		MessageId int64           `json:"id"`
	}
)
