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
	// Message
	MessageObject struct {
		UserId  int64  `json:"userid"`
		Message string `json:"msg"`
	}
)

type (
	// User Info
	UserInfo struct {
		Id   int64  `json:"id"`
		Name string `json:"name"`
	}
)
