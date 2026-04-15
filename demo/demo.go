package demo

import (
	windows "github.com/shileias/gdesk/command/windows"
)

type UserInfo struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// 调试的操作窗口
func debug() {
	a := windows.GetApp()
	a.MainWindow().Hide()
}

//gdesk:bind Example
func Example(name string) (string, error) {
	return "Hello from plugin demo, " + name, nil
}

//gdesk:bind Add
func Add(a, b int) (int, error) {
	return a + b, nil
}

//gdesk:bind Multiply
func Multiply(a, b int) (int, error) {
	return a * b, nil
}

//gdesk:bind GetUser
func GetUser(id int) (UserInfo, error) {
	return UserInfo{
		ID:    id,
		Name:  "Demo User",
		Email: "demo@example.com",
	}, nil
}

//gdesk:bind GetUsers
func GetUsers() ([]UserInfo, error) {
	return []UserInfo{
		{ID: 1, Name: "Alice", Email: "alice@example.com"},
		{ID: 2, Name: "Bob", Email: "bob@example.com"},
		{ID: 3, Name: "Charlie", Email: "charlie@example.com"},
	}, nil
}

func init() {
	windows.RegisterPluginFunctions("plugins/demo",
		"Example", Example,
		"Add", Add,
		"Multiply", Multiply,
		"GetUser", GetUser,
		"GetUsers", GetUsers,
	)
}
