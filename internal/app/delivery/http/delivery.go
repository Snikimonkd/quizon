package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"quizon_bot/internal/pkg/logger"
)

type Usecase interface {
	RegisterUsecase
	RegistrationsUsecase
	RegisterAvailableUsecase
	// LoginUsecase
}

type delivery struct {
	usecase Usecase
}

func NewDelivery(usecase Usecase) *delivery {
	return &delivery{
		usecase: usecase,
	}
}

type Error struct {
	Msg string `json:"msg"`
}

func ResponseWithJSON(w http.ResponseWriter, code int, body interface{}) {
	w.WriteHeader(code)
	if body == nil {
		return
	}

	err := json.NewEncoder(w).Encode(body)
	if err != nil {
		logger.Errorf("can't write body: %v", err)
	}
}

func UnmarshalRequest(body io.ReadCloser, value interface{}) error {
	decoder := json.NewDecoder(body)
	err := decoder.Decode(value)
	if err != nil {
		return fmt.Errorf("can't unmarshal request: %w", err)
	}

	return nil
}
