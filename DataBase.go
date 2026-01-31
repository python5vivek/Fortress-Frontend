package main

import (
	"errors"
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

type USERS struct {
	gorm.Model
	UserID    int
	Username  string
	FirstName string
	LastName  string
}

type MESSAGE struct {
	gorm.Model
	From      string
	To        string
	Message   string
	ChatterID int
}

type ChatLister struct {
	gorm.Model
	ChatterID int
	Username  string
	FirstName string
}

func initDB() {
	db, err := gorm.Open(sqlite.Open("app.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database:", err)
	}
	DB = db
	DB.AutoMigrate(
		&USERS{},
		&MESSAGE{},
		&ChatLister{},
	)
}

func AddUser(id int, username string, fn string, ln string) {
	user := USERS{UserID: id, Username: username, FirstName: fn, LastName: ln}
	DB.Create(&user)

}

func GetSpecificUser(id int) *USERS {
	var user USERS
	DB.Where("user_id = ?", id).Find(&user)
	return &user
}

func GetMessages(id int) []MESSAGE {
	var messages []MESSAGE
	DB.Where("chatter_id = ?", id).Find(&messages)
	return messages
}

func GetChattedList() []ChatLister {
	var chatlist []ChatLister
	DB.Find(&chatlist)
	return chatlist
}

func AddMessage(chatterID int, to string, from string, message string) {
	newchat := MESSAGE{ChatterID: chatterID, To: to, From: from, Message: message}
	DB.Create(&newchat)
}

func CheckIsIt(id int) bool {
	var chat ChatLister
	result := DB.Where("chatter_id = ?", id).First(&chat)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return false
	}
	return result.Error == nil
}

func CreateNewChatLis(id int, username string, firstname string) {
	DB.Create(&ChatLister{ChatterID: id, Username: username, FirstName: firstname})
}
