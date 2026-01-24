package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/gorilla/websocket"
)

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

func AllUsers(token string) []interface{} {
	url := BaseURL + "/users/"
	resp, err := getJSON(url, token)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	users := result["users"].([]interface{})
	return users
}
