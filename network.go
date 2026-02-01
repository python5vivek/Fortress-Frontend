package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

type User struct {
	ID         int    `json:"id"`
	Username   string `json:"username"`
	First_Name string `json:"first_name"`
	Last_Name  string `json:"last_name"`
}

func ConnectWS(baseURL, token string) (*websocket.Conn, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	q := u.Query()
	q.Set("token", token)
	u.RawQuery = q.Encode()

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
func SendToUser(conn *websocket.Conn, userID int, message string) error {
	payload := map[string]interface{}{
		"to":      "user",
		"id":      userID,
		"message": message,
	}
	return conn.WriteJSON(payload)
}
func SendToGlobal(conn *websocket.Conn, message string) error {
	payload := map[string]interface{}{
		"to":      "global",
		"message": message,
	}
	return conn.WriteJSON(payload)
}
func ReceiveMessage(conn *websocket.Conn) (map[string]interface{}, error) {
	_, data, err := conn.ReadMessage()
	if err != nil {
		return nil, err
	}

	var msg map[string]interface{}
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, err
	}

	return msg, nil
}

func postJSON(url string, data map[string]string) (*http.Response, error) {
	body, _ := json.Marshal(data)
	return http.Post(
		url,
		"application/json",
		bytes.NewBuffer(body),
	)
}

func getJSON(url, token string) (*http.Response, error) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Token "+token)
	return client.Do(req)
}

func IsOnline(url string) bool {
	client := &http.Client{Timeout: 5 * time.Second}
	req, _ := http.NewRequest("GET", url, nil)
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusNotFound
}

func AllUsers(token string) ([]User, error) {
	url := BaseURL + "/users/"
	resp, err := getJSON(url, token)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var users []User
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return nil, err
	}

	return users, nil
}
