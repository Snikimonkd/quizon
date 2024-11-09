package model

type Register struct {
	GameID      int64   `json:"game_id"`
	TgContact   string  `json:"telegram"`
	TeamID      *string `json:"team_id,omitempty"`
	TeamName    string  `json:"team_name"`
	CaptainName string  `json:"captain_name"`
	Phone       string  `json:"phone"`
	GroupName   *string `json:"group_name"`
	TeamSize    string  `json:"team_size"`
}

type RegisterAvailableRequest struct {
	GameID int64 `json:"game_id"`
}

type Registration struct {
	Number       int64   `json:"number"`
	Telegram     string  `json:"tg_contact"`
	TeamID       *string `json:"team_id,omitempty"`
	TeamName     string  `json:"team_name"`
	CaptainName  string  `json:"captain_name"`
	Phone        string  `json:"phone"`
	GroupName    *string `json:"group_name"`
	TeamSize     string  `json:"amount"`
	RegisteredAt string  `json:"registered_at"`
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
