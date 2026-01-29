package main

import (
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
	DB.Where("user_id = ?", id).First(&user)
	return &user
}
