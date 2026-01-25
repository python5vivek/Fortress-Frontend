package main

import (
	"encoding/json"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/gorilla/websocket"
)

var wsConn *websocket.Conn
var messageBox *fyne.Container
var w *fyne.Window

func ConnectingPage() fyne.CanvasObject {
	loader := widget.NewActivity()
	loader.Start()

	label := widget.NewLabelWithStyle(
		"Connecting ....",
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	content := container.NewVBox(
		loader,
		label,
	)

	return container.New(
		layout.NewCenterLayout(),
		container.NewPadded(content),
	)
}

func LoginView(w fyne.Window, switchToRegister func()) fyne.CanvasObject {
	username := widget.NewEntry()
	username.SetPlaceHolder("Username")
	password := widget.NewPasswordEntry()
	password.SetPlaceHolder("Password")

	status := widget.NewLabel("")

	loginBtn := widget.NewButton("Login", func() {
		resp, err := postJSON(
			LoginURL,
			map[string]string{
				"username": username.Text,
				"password": password.Text,
			},
		)

		if err != nil {
			status.SetText("Server error")
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			status.SetText("Invalid credentials")
			return
		}
		var respData map[string]string
		json.NewDecoder(resp.Body).Decode(&respData)
		token, ok := respData["token"]
		if !ok {
			status.SetText("Invalid server response")
			return
		}
		status.SetText("Login successful")
		SaveToken(token)
		conn, err := ConnectWS(WSURL, token)
		if err != nil {
			status.SetText("Failed to connect to WebSocket")
			return
		}
		wsConn = conn
		HomePage(w)
	})

	registerBtn := widget.NewButton("Register", switchToRegister)

	form := container.NewVBox(
		widget.NewLabelWithStyle("Login", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		username,
		password,
		loginBtn,
		registerBtn,
		status,
	)
	return form
}
func RegisterView(w fyne.Window, switchToLogin func()) fyne.CanvasObject {
	username := widget.NewEntry()
	first := widget.NewEntry()
	last := widget.NewEntry()
	email := widget.NewEntry()
	password := widget.NewPasswordEntry()
	confirm := widget.NewPasswordEntry()

	username.SetPlaceHolder("Username")
	first.SetPlaceHolder("First name")
	last.SetPlaceHolder("Last name")
	email.SetPlaceHolder("Email")
	password.SetPlaceHolder("Password")
	confirm.SetPlaceHolder("Confirm password")

	status := widget.NewLabel("")

	registerBtn := widget.NewButton("Create Account", func() {
		if password.Text != confirm.Text {
			status.SetText("Passwords do not match")
			return
		}

		resp, err := postJSON(
			RegisterURL,
			map[string]string{
				"username":   username.Text,
				"first_name": first.Text,
				"last_name":  last.Text,
				"email":      email.Text,
				"password":   password.Text,
			},
		)

		if err != nil {
			status.SetText("Server error")
			return
		}
		if resp.StatusCode != 201 {
			status.SetText("Invalid credentials")
			return
		}
		var respData map[string]string
		json.NewDecoder(resp.Body).Decode(&respData)
		token, ok := respData["token"]
		if !ok {
			status.SetText("Invalid server response")
			return
		}
		status.SetText("Login successful")
		SaveToken(token)
		conn, err := ConnectWS(WSURL, token)
		if err != nil {
			status.SetText("Failed to connect to WebSocket")
			return
		}
		wsConn = conn
		HomePage(w)
	})

	backBtn := widget.NewButton("Back to Login", switchToLogin)

	return container.NewVBox(
		widget.NewLabelWithStyle("Register", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		username,
		container.NewGridWithColumns(2, first, last),
		email,
		password,
		confirm,
		registerBtn,
		backBtn,
		status,
	)
}

func ThemeSwitch() fyne.CanvasObject {
	settings := LoadSettings()

	check := widget.NewCheck("Dark Mode", func(on bool) {
		if on {
			fyne.CurrentApp().Settings().SetTheme(theme.DarkTheme())
			settings.Theme = "dark"
		} else {
			fyne.CurrentApp().Settings().SetTheme(theme.LightTheme())
			settings.Theme = "light"
		}
		SaveSettings(settings)
	})

	if settings.Theme == "dark" {
		check.SetChecked(true)
	}

	return container.NewHBox(
		layout.NewSpacer(),
		check,
	)
}

func ProfileInfo(username, first, last string, onLogout func()) fyne.CanvasObject {
	userLbl := widget.NewLabelWithStyle(
		username,
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	nameLbl := widget.NewLabel(
		fmt.Sprintf("%s %s", first, last),
	)
	nameLbl.Alignment = fyne.TextAlignCenter

	logoutBtn := widget.NewButtonWithIcon(
		"Logout",
		theme.LogoutIcon(),
		onLogout,
	)

	return container.NewVBox(
		layout.NewSpacer(),
		userLbl,
		nameLbl,
		layout.NewSpacer(),
		logoutBtn,
		layout.NewSpacer(),
	)
}

func ProfilePage(
	w fyne.Window,
	username, first, last string,
	onLogout func(),
) fyne.CanvasObject {

	return container.NewBorder(
		container.NewPadded(ThemeSwitch()), // TOP (right aligned)
		nil,
		nil,
		nil,
		container.NewCenter(
			ProfileInfo(username, first, last, onLogout),
		),
	)
}

func MessageBubble(username, text string) fyne.CanvasObject {
	return container.NewVBox(
		widget.NewLabelWithStyle(
			username,
			fyne.TextAlignLeading,
			fyne.TextStyle{Bold: true},
		),
		widget.NewLabel(text),
	)
}

func ChatHeader(w fyne.Window, username string, onBack func()) fyne.CanvasObject {
	backBtn := widget.NewButtonWithIcon(
		"",
		theme.NavigateBackIcon(),
		onBack,
	)

	title := widget.NewLabelWithStyle(
		username,
		fyne.TextAlignLeading,
		fyne.TextStyle{Bold: true},
	)

	return container.NewPadded(
		container.NewHBox(
			backBtn,
			title,
			layout.NewSpacer(),
		),
	)
}

func ChatMessages() (*fyne.Container, *container.Scroll) {
	box := container.NewVBox()
	scroll := container.NewVScroll(box)
	scroll.ScrollToBottom()
	return box, scroll
}

func ChatBubble(text string, isMe bool) fyne.CanvasObject {
	label := widget.NewLabel(text)
	//label.Wrapping = fyne.TextWrapWord
	label.Resize(fyne.NewSize(200, label.MinSize().Height))

	bg := canvas.NewRectangle(theme.InputBackgroundColor())

	bubble := container.NewPadded(
		container.NewMax(bg, label),
	)

	if isMe {
		return container.NewHBox(layout.NewSpacer(), bubble)
	}
	return container.NewHBox(bubble, layout.NewSpacer())
}

func ChatInput(onSend func(msg string)) fyne.CanvasObject {
	entry := widget.NewEntry()
	entry.SetPlaceHolder("Write message...")

	sendBtn := widget.NewButtonWithIcon(
		"",
		theme.MailSendIcon(),
		func() {
			if entry.Text == "" {
				return
			}
			onSend(entry.Text)
			entry.SetText("")
		},
	)

	entry.OnSubmitted = func(_ string) {
		sendBtn.OnTapped()
	}

	return container.NewBorder(
		nil, nil,
		sendBtn, // LEFT
		nil,
		entry,
	)
}

func ChatPage(w fyne.Window, username string, onBack func()) fyne.CanvasObject {
	msgBox, scroll := ChatMessages()

	// TEMP demo messages
	msgBox.Add(ChatBubble("Hello!", false))
	msgBox.Add(ChatBubble("Hi, how are you?", true))
	msgBox.Add(ChatBubble("All good 👍", false))

	input := ChatInput(func(msg string) {
		msgBox.Add(ChatBubble(msg, true))
		scroll.ScrollToBottom()
	})

	return container.NewBorder(
		ChatHeader(w, username, onBack), // TOP
		input,                           // BOTTOM
		nil,
		nil,
		scroll, // CENTER
	)
}

func ChatRow(username, lastMsg, time string, onClick func()) fyne.CanvasObject {
	name := widget.NewLabelWithStyle(
		username,
		fyne.TextAlignLeading,
		fyne.TextStyle{Bold: true},
	)

	msg := widget.NewLabel(lastMsg)
	msg.Truncation = fyne.TextTruncateEllipsis

	timestamp := widget.NewLabel(time)
	//timestamp.Alignment = fyne.TextAlignTrailing

	left := container.NewVBox(name, timestamp)

	button := widget.NewButton("", onClick)
	button.Text = "Chat"
	button.Importance = widget.HighImportance
	pressbu := container.NewHBox(layout.NewSpacer(), button, layout.NewSpacer())
	return container.NewBorder(
		nil,
		nil,
		nil,
		container.NewVBox(layout.NewSpacer(), pressbu, layout.NewSpacer()),
		left,
	)
}

func ChatList(w fyne.Window) fyne.CanvasObject {
	list := container.NewVBox()

	// MOCK DATA (replace later)
	chats := []struct {
		User string
		Last string
		Time string
	}{
		{"alice", "Hi there", "10:30"},
		{"bob", "See you", "09:15"},
		{"charlie", "Ok", "Yesterday"},
	}

	for _, c := range chats {
		row := ChatRow(c.User, c.Last, c.Time, func() {
			w.SetContent(ChatPage(w, c.User, func() {
				HomePage(w)
			}))
		})
		list.Add(row)
	}

	return container.NewVScroll(list)
}

func ChatAppBar(w fyne.Window) fyne.CanvasObject {
	menu := fyne.NewMenu("Menu",
		fyne.NewMenuItem("Get All Users", func() {
			w.SetContent(
				AllUsersPage(w, func() {
					HomePage(w)
				}),
			)
		}),
		fyne.NewMenuItem("Search User", func() {
			dialog.ShowInformation("Search", "Search by username (TODO)", w)
		}),
		fyne.NewMenuItem("About", func() {
			dialog.ShowInformation("About", "Chat App v1.0", w)
		}),
	)

	menuBtn := widget.NewButtonWithIcon(
		"",
		theme.MenuIcon(),
		func() {
			widget.ShowPopUpMenuAtPosition(
				menu,
				w.Canvas(),
				fyne.NewPos(10, 50),
			)
		},
	)

	title := widget.NewLabelWithStyle(
		"Chats",
		fyne.TextAlignLeading,
		fyne.TextStyle{Bold: true},
	)

	bar := container.NewHBox(
		menuBtn,
		title,
		layout.NewSpacer(),
	)

	return container.NewPadded(bar)
}

func ChatHomePage(w fyne.Window) fyne.CanvasObject {
	return container.NewBorder(
		ChatAppBar(w), // TOP only
		nil,
		nil,
		nil,
		ChatList(w),
	)
}

func GroupPage() fyne.CanvasObject {
	return container.NewCenter(widget.NewLabel("Group Page"))
}

func GlobalChatPage() fyne.CanvasObject {
	Title := widget.NewLabelWithStyle(
		"Global Chat",
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)
	messageBox = container.NewVBox()
	scroll := container.NewVScroll(messageBox)
	scroll.SetMinSize(fyne.NewSize(0, 400))
	input := widget.NewEntry()
	input.SetPlaceHolder("Type a message...")
	sendBtn := widget.NewButton("Send", func() {
		if input.Text == "" {
			return
		}
		err := SendToGlobal(wsConn, input.Text)
		if err != nil {
			fmt.Println("Error sending message:", err)
			return
		}
		input.SetText("")
		scroll.ScrollToBottom()
	})
	sendBtn.Importance = widget.HighImportance
	inputBar := container.NewBorder(
		nil,
		nil,
		nil,
		sendBtn,
		input,
	)
	return container.NewBorder(
		Title,
		inputBar,
		nil,
		nil,
		scroll,
	)
}

func AllUsersHeader(w fyne.Window, onBack func()) fyne.CanvasObject {
	backBtn := widget.NewButtonWithIcon(
		"",
		theme.NavigateBackIcon(),
		onBack,
	)

	title := widget.NewLabelWithStyle(
		"All Users",
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	return container.NewPadded(
		container.NewHBox(
			backBtn,
			layout.NewSpacer(),
			title,
			layout.NewSpacer(),
		),
	)
}

func UserRow(
	username, firstName, lastName string,
	onChat func(),
) fyne.CanvasObject {

	usernameLbl := widget.NewLabelWithStyle(
		username,
		fyne.TextAlignLeading,
		fyne.TextStyle{Bold: true},
	)

	nameLbl := widget.NewLabel(
		fmt.Sprintf("%s %s", firstName, lastName),
	)

	info := container.NewVBox(
		usernameLbl,
		nameLbl,
	)

	chatBtn := widget.NewButton("Chat", onChat)
	chatBtn.Importance = widget.HighImportance

	return container.NewPadded(
		container.NewHBox(
			info,               // LEFT (username + name)
			layout.NewSpacer(), // PUSHES button to right
			container.NewVBox(layout.NewSpacer(), chatBtn, layout.NewSpacer()), // RIGHT
		),
	)
}

func AllUsersList(w fyne.Window) fyne.CanvasObject {
	list := container.NewVBox()
	token, _ := GetToken()
	users, _ := AllUsers(token)
	// MOCK DATA

	for _, u := range users {
		row := UserRow(
			u.Username,
			u.First_Name,
			u.Last_Name,
			func() {
				dialog.ShowInformation(
					"Chat",
					"Open chat with "+u.Username,
					w,
				)
			},
		)
		list.Add(row)
	}

	return container.NewVScroll(list)
}

func AllUsersPage(w fyne.Window, onBack func()) fyne.CanvasObject {
	return container.NewBorder(
		AllUsersHeader(w, onBack), // TOP
		nil,
		nil,
		nil,
		AllUsersList(w), // CENTER
	)
}

func HomePage(w fyne.Window) {
	tabs := container.NewAppTabs(
		container.NewTabItemWithIcon(
			"Chat",
			theme.HomeIcon(),
			ChatHomePage(w),
		),
		container.NewTabItemWithIcon(
			"Group",
			theme.AccountIcon(),
			GroupPage(),
		),
		container.NewTabItemWithIcon(
			"Global",
			theme.SearchIcon(),
			GlobalChatPage(),
		),
		container.NewTabItemWithIcon(
			"Profile",
			theme.SettingsIcon(),
			ProfilePage(w, "username", "First", "Last", func() {
				ClearToken()
			}),
		),
	)
	tabs.SetTabLocation(container.TabLocationBottom)
	w.SetContent(tabs)
}

func main() {
	//ClearToken()
	a := app.New()
	w := a.NewWindow("Chat App")

	settings := LoadSettings()

	if settings.Theme == "dark" {
		a.Settings().SetTheme(theme.DarkTheme())
	} else {
		a.Settings().SetTheme(theme.LightTheme())
	}

	token, has := GetToken()
	if !has {
		token = ""
	}
	w.Resize(fyne.NewSize(400, 500))
	var showLogin func()
	var showRegister func()
	showLogin = func() {
		w.SetContent(LoginView(w, showRegister))
	}
	showRegister = func() {
		w.SetContent(RegisterView(w, showLogin))
	}
	if HasToken() {
		w.SetContent(ConnectingPage())

		go func() {
			conn, err := ConnectWS(WSURL, token)
			wsConn = conn

			fyne.Do(func() {
				if err != nil {
					showLogin()
					return
				}

				HomePage(w)

				go func() {
					for {
						msg, err := ReceiveMessage(conn)
						if err != nil {
							fmt.Println("WebSocket closed:", err)
							return
						}
						fyne.Do(func() {
							msg := MessageBubble(msg["who"].(string), msg["message"].(string))
							messageBox.Add(msg)
						})
					}
				}()

			})
		}()
	} else {
		showLogin()
	}
	w.ShowAndRun()
}
