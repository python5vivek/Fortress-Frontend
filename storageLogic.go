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
	UserName    string        `json:"user_name"`
	FirstName   string        `json:"first_name"`
	LastName    string        `json:"last_name"`
	ChannelName string        `json:"channel_name"`
	Messages    []ChatMessage `json:"messages"`
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
		UserName:    username,
		FirstName:   first,
		LastName:    last,
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
