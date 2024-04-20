package model

import "time"

type Registration struct {
	ID              int64
	GameID          int64
	TeamID          string
	CaptainName     string
	CaptainGroup    string
	CaptainTelegram string
	TeamName        string
	TeamSize        int64
	CreatedAt       time.Time
}

type Modal struct {
	Header string
	Text   string
	Button string
}

type RegistrationForm struct {
	Registration *Registration
	Modal        *Modal
}
