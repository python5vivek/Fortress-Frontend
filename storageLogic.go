package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type ChatMessage struct {
	From      string `json:"from"`
	To        string `json:"to"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

type ChatFile struct {
	Id          int
	UserName    string        `json:"user_name"`
	First_Name  string        `json:"first_name"`
	Last_Name   string        `json:"last_name"`
	ChannelName string        `json:"channel_name"`
	Messages    []ChatMessage `json:"messages"`
}

type ChattedUser struct {
	ID          int
	UserName    string
	FirstName   string
	LastName    string
	ChannelName string
}

func chatDir() string {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".fortresschat", "chats")
	os.MkdirAll(dir, 0700)
	return dir
}

func chatFilePath(userID int) string {
	return filepath.Join(chatDir(), fmt.Sprintf("user_%d.json", userID))
}

func LoadOrCreateChat(
	userID int,
	username, first, last string,
) (*ChatFile, error) {

	path := chatFilePath(userID)

	// if exists → load
	if data, err := os.ReadFile(path); err == nil {
		var chat ChatFile
		if err := json.Unmarshal(data, &chat); err != nil {
			return nil, err
		}
		return &chat, nil
	}

	// else → create
	chat := ChatFile{
		Id:          userID,
		UserName:    username,
		First_Name:  first,
		Last_Name:   last,
		ChannelName: fmt.Sprintf("user_%d", userID),
		Messages:    []ChatMessage{},
	}

	if err := SaveChat(userID, &chat); err != nil {
		return nil, err
	}

	return &chat, nil
}

func SaveChat(userID int, chat *ChatFile) error {
	data, err := json.MarshalIndent(chat, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(chatFilePath(userID), data, 0600)
}

func AppendMessage(
	userID int,
	chat *ChatFile,
	from, to, message string,
) error {

	msg := ChatMessage{
		From:      from,
		To:        to,
		Message:   message,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	chat.Messages = append(chat.Messages, msg)
	return SaveChat(userID, chat)
}

func GetMessages(userID int) ([]ChatMessage, error) {
	data, err := os.ReadFile(chatFilePath(userID))
	if err != nil {
		return nil, err
	}

	var chat ChatFile
	if err := json.Unmarshal(data, &chat); err != nil {
		return nil, err
	}

	return chat.Messages, nil
}

func GetChattedUsers() ([]ChattedUser, error) {
	var users []ChattedUser

	files, err := os.ReadDir(chatDir())
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// extract user ID from filename: user_<id>.json
		var id int
		_, err := fmt.Sscanf(file.Name(), "user_%d.json", &id)
		if err != nil {
			continue
		}

		data, err := os.ReadFile(filepath.Join(chatDir(), file.Name()))
		if err != nil {
			continue
		}

		var chat ChatFile
		if err := json.Unmarshal(data, &chat); err != nil {
			continue
		}

		users = append(users, ChattedUser{
			ID:          id,
			UserName:    chat.UserName,
			FirstName:   chat.First_Name,
			LastName:    chat.Last_Name,
			ChannelName: chat.ChannelName,
		})
	}

	return users, nil
}

func GetChatByUserID(userID int) (*ChatFile, error) {
	path := chatFilePath(userID)

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var chat ChatFile
	if err := json.Unmarshal(data, &chat); err != nil {
		return nil, err
	}

	return &chat, nil
}
