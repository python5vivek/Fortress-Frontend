package main

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/storage"
)

type ChatMessage struct {
	From      string `json:"from"`
	To        string `json:"to"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

type ChatFile struct {
	Id          int           `json:"id"`
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

//
// helpers
//

func ensureChatsDir(app fyne.App) error {
	root := app.Storage().RootURI()

	chatsDir, err := storage.Child(root, "chats")
	if err != nil {
		return err
	}

	keepFile, err := storage.Child(chatsDir, ".keep")
	if err != nil {
		return err
	}

	// 🔑 Just write. Do NOT read.
	w, err := storage.Writer(keepFile)
	if err != nil {
		return err
	}
	w.Close()

	return nil
}
func chatDirURI(app fyne.App) (fyne.URI, error) {
	return app.Storage().RootURI(), nil
}

func chatFileURI(app fyne.App, userID int) (fyne.URI, error) {
	dir, err := chatDirURI(app)
	if err != nil {
		return nil, err
	}
	return storage.Child(dir, fmt.Sprintf("user_%d.json", userID))
}

//
// core functions
//

func LoadOrCreateChat(
	app fyne.App,
	userID int,
	username, first, last string,
) (*ChatFile, error) {

	chatsDir := app.Storage().RootURI()

	chatURI, err := storage.Child(
		chatsDir,
		fmt.Sprintf("user_%d.json", userID),
	)
	if err != nil {
		return nil, err
	}

	// 1️⃣ Try reading
	if r, err := storage.Reader(chatURI); err == nil {
		defer r.Close()

		data, _ := io.ReadAll(r)
		var chat ChatFile
		if err := json.Unmarshal(data, &chat); err != nil {
			return nil, err
		}
		return &chat, nil
	}

	// 2️⃣ Create file if missing
	chat := ChatFile{
		Id:          userID,
		UserName:    username,
		First_Name:  first,
		Last_Name:   last,
		ChannelName: fmt.Sprintf("user_%d", userID),
		Messages:    []ChatMessage{},
	}

	w, err := storage.Writer(chatURI)
	if err != nil {
		return nil, err
	}
	defer w.Close()

	data, _ := json.MarshalIndent(chat, "", "  ")
	_, err = w.Write(data)
	if err != nil {
		return nil, err
	}

	return &chat, nil
}

func SaveChat(app fyne.App, userID int, chat *ChatFile) error {
	uri, err := chatFileURI(app, userID)
	if err != nil {
		return err
	}

	w, err := storage.Writer(uri)
	if err != nil {
		return err
	}
	defer w.Close()

	data, err := json.MarshalIndent(chat, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println("Saving chat file for user:", userID)
	fmt.Println(string(data))
	_, err = w.Write(data)
	return err
}

func AppendMessage(
	app fyne.App,
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

	dir := app.Storage().RootURI()

	uri, err := storage.Child(dir, fmt.Sprintf("user_%d.json", userID))
	if err != nil {
		return err
	}

	w, err := storage.Writer(uri)
	if err != nil {
		return err
	}
	defer w.Close()

	data, err := json.MarshalIndent(chat, "", "  ")
	if err != nil {
		return err
	}

	_, err = w.Write(data)
	return err
}

func GetMessages(app fyne.App, userID int) ([]ChatMessage, error) {
	uri, err := chatFileURI(app, userID)
	if err != nil {
		return nil, err
	}

	r, err := storage.Reader(uri)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	data, _ := io.ReadAll(r)

	var chat ChatFile
	if err := json.Unmarshal(data, &chat); err != nil {
		return nil, err
	}

	return chat.Messages, nil
}

func GetChatByUserID(app fyne.App, userID int) (*ChatFile, error) {
	uri, err := chatFileURI(app, userID)
	if err != nil {
		return nil, err
	}

	r, err := storage.Reader(uri)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	data, _ := io.ReadAll(r)

	var chat ChatFile
	if err := json.Unmarshal(data, &chat); err != nil {
		return nil, err
	}

	return &chat, nil
}

func GetChattedUsers(app fyne.App) ([]ChattedUser, error) {
	var users []ChattedUser

	dir, err := chatDirURI(app)
	if err != nil {
		return users, nil
	}

	list, err := storage.List(dir)
	if err != nil {
		return users, nil
	}

	for _, uri := range list {
		r, err := storage.Reader(uri)
		if uri.Name() == "auth.json" {
			continue
		}
		if uri.Name() == "settings.json" {
			continue
		}
		if uri.Name() == "preferences.json" {
			continue
		}
		if err != nil {
			continue
		}

		data, _ := io.ReadAll(r)
		r.Close()

		var chat ChatFile
		if err := json.Unmarshal(data, &chat); err != nil {
			continue
		}

		users = append(users, ChattedUser{
			ID:          chat.Id,
			UserName:    chat.UserName,
			FirstName:   chat.First_Name,
			LastName:    chat.Last_Name,
			ChannelName: chat.ChannelName,
		})
	}

	return users, nil
}
