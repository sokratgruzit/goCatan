package models

type User struct {
	ID          int64  `json:"id"`
	Email       string `json:"email"`
	Password    string `json:"-"`
	Username    string `json:"username"`
	Balance     int64  `json:"balance"`
	DemoBalance int64  `json:"demo_balance"`
}
