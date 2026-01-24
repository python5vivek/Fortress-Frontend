package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func TokenFilePath() string {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".chatapp")
	os.MkdirAll(dir, 0700)
	return filepath.Join(dir, "auth.json")
}
func SaveToken(token string) error {
	data := map[string]string{
		"token": token,
	}

	bytes, _ := json.Marshal(data)
	return os.WriteFile(TokenFilePath(), bytes, 0600)
}
func HasToken() bool {
	_, err := os.Stat(TokenFilePath())
	return err == nil
}
func GetToken() (string, bool) {
	bytes, err := os.ReadFile(TokenFilePath())
	if err != nil {
		return "", false
	}

	var data map[string]string
	if err := json.Unmarshal(bytes, &data); err != nil {
		return "", false
	}

	token, ok := data["token"]
	return token, ok
}
func ClearToken() {
	os.Remove(TokenFilePath())
}
