package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func settingsFilePath() string {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".fortresschat")
	os.MkdirAll(dir, 0700)
	return filepath.Join(dir, "settings.json")
}

type AppSettings struct {
	Theme string `json:"theme"` // "dark" or "light"
}

func LoadSettings() AppSettings {
	path := settingsFilePath()

	data, err := os.ReadFile(path)
	if err != nil {
		// default settings
		return AppSettings{
			Theme: "light",
		}
	}

	var s AppSettings
	if err := json.Unmarshal(data, &s); err != nil {
		return AppSettings{Theme: "light"}
	}

	return s
}

func SaveSettings(s AppSettings) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(settingsFilePath(), data, 0600)
}
