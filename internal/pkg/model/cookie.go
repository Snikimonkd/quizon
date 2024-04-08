package model

import (
	"time"

	"github.com/google/uuid"
)

type Cookie struct {
	AdminName string
	Value     uuid.UUID
	Expires   time.Time
}
