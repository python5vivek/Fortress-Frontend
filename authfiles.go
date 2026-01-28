package main

import (
	"encoding/json"
	"io"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/storage"
)

// helper
func tokenURI(app fyne.App) (fyne.URI, error) {
	return storage.Child(app.Storage().RootURI(), "auth.json")
}

func SaveToken(app fyne.App, token string) error {
	uri, err := tokenURI(app)
	if err != nil {
		return err
	}

	w, err := storage.Writer(uri)
	if err != nil {
		return err
	}
	defer w.Close()

	data := map[string]string{
		"token": token,
	}

	b, err := json.Marshal(data)
	if err != nil {
		return err
	}

	_, err = w.Write(b)
	return err
}

func HasToken(app fyne.App) bool {
	uri, err := tokenURI(app)
	if err != nil {
		return false
	}

	r, err := storage.Reader(uri)
	if err != nil {
		return false
	}
	r.Close()
	return true
}

func GetToken(app fyne.App) (string, bool) {
	uri, err := tokenURI(app)
	if err != nil {
		return "", false
	}

	r, err := storage.Reader(uri)
	if err != nil {
		return "", false
	}
	defer r.Close()

	b, err := io.ReadAll(r)
	if err != nil {
		return "", false
	}

	var data map[string]string
	if err := json.Unmarshal(b, &data); err != nil {
		return "", false
	}

	token, ok := data["token"]
	return token, ok
}

func ClearToken(app fyne.App) {
	uri, err := tokenURI(app)
	if err != nil {
		return
	}

	// overwrite with empty file (Delete is not guaranteed)
	w, err := storage.Writer(uri)
	if err != nil {
		return
	}
	w.Close()
}
