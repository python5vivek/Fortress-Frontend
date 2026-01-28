package main

import (
	"encoding/json"
	"io"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/storage"
)

type AppSettings struct {
	Theme string `json:"theme"` // "dark" or "light"
}

// helper: settings.json URI
func settingsURI(app fyne.App) (fyne.URI, error) {
	return storage.Child(app.Storage().RootURI(), "settings.json")
}

func LoadSettings(app fyne.App) AppSettings {
	uri, err := settingsURI(app)
	if err != nil {
		return AppSettings{Theme: "light"}
	}

	r, err := storage.Reader(uri)
	if err != nil {
		// default settings
		return AppSettings{Theme: "light"}
	}
	defer r.Close()

	data, err := io.ReadAll(r)
	if err != nil {
		return AppSettings{Theme: "light"}
	}

	var s AppSettings
	if err := json.Unmarshal(data, &s); err != nil {
		return AppSettings{Theme: "light"}
	}

	// safety default
	if s.Theme == "" {
		s.Theme = "light"
	}

	return s
}

func SaveSettings(app fyne.App, s AppSettings) error {
	uri, err := settingsURI(app)
	if err != nil {
		return err
	}

	w, err := storage.Writer(uri)
	if err != nil {
		return err
	}
	defer w.Close()

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	_, err = w.Write(data)
	return err
}
