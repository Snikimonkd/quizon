package model

import "time"

type Registration struct {
	ID              int64
	GameID          int64
	TeamID          int64
	CaptainName     string
	CaptainGroup    string
	CaptainTelegram string
	TeamName        string
	TeamSize        int64
	CreatedAt       time.Time
}
