package model

import "time"

type CreateGameRequest struct {
	StartTime            time.Time `json:"start_time"`
	Location             string    `json:"location"`
	Name                 string    `json:"name"`
	MainAmount           int64     `json:"main_amount"`
	ReserveAmount        int64     `json:"reserve_amount"`
	RegistartionOpenTime time.Time `json:"registartion_open_time"`
}

type Register struct {
	GameID        int64   `json:"game_id"`
	TgContact     string  `json:"telegram"`
	TeamID        *string `json:"team_id,omitempty"`
	TeamName      string  `json:"team_name"`
	CaptainName   string  `json:"captain_name"`
	Phone         string  `json:"phone"`
	GroupName     *string `json:"group_name"`
	PlayersAmount int64   `json:"players_amount"`
}

type RegisterAvailableRequest struct {
	GameID int64 `json:"game_id"`
}

type Registration struct {
	Number        int64   `json:"number"`
	Telegram      string  `json:"tg_contact"`
	TeamID        *string `json:"team_id,omitempty"`
	TeamName      string  `json:"team_name"`
	CaptainName   string  `json:"captain_name"`
	Phone         string  `json:"phone"`
	GroupName     *string `json:"group_name"`
	PlayersAmount int64   `json:"players_amount"`
	RegisteredAt  string  `json:"registered_at"`
}

type RegisterAvailableResponse struct {
	Available RegistrationStatus `json:"available"`
}

type ListRegistrationsRequest struct {
	GameID int64 `json:"game_id"`
}

type RegistrationStatus string

const (
	Available    RegistrationStatus = "available"
	Reserve      RegistrationStatus = "reserve"
	Closed       RegistrationStatus = "closed"
	NotOpenedYet RegistrationStatus = "not_opened_yet"
)

type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type Game struct {
	ID                   int64     `json:"id"`
	CreatedAt            time.Time `json:"created_at"`
	StartTime            time.Time `json:"start_time"`
	Location             string    `json:"location"`
	Name                 string    `json:"name"`
	MainAmount           int64     `json:"main_amount"`
	ReserveAmount        int64     `json:"reserve_amount"`
	RegistartionOpenTime time.Time `json:"registartion_open_time"`
}
