package models

type User struct {
	ID            int64  `json:"id"`
	Email         string `json:"email"`
	Password      string `json:"-"`
	Username      string `json:"username"`
	Balance       int64  `json:"balance"`
	DemoBalance   int64  `json:"demoBalance"`
	Address       string `json:"address"`
	AccessToken   string `json:"access_token"`
	Roles         string `json:"roles"`
	Avatar        string `json:"avatar"`
	GameStarted   bool   `json:"gameStarted"`
	SwitchAccount bool   `json:"switchAccount"`
}
