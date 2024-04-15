package model

import "time"

type Game struct {
	ID                int64
	Location          string
	StartTime         time.Time
	Name              string
	TeamsAmount       int64
	Reserve           int64
	RegistrationStart time.Time
	RegisteredTeams   int64
	Comment           string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
