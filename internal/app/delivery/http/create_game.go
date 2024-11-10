package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	httpModel "quizon/internal/app/delivery/http/model"
)

func kek() {
	t := time.Now()
	b, _ := json.Marshal(t)
	fmt.Println(string(b))
}

type CreateGameUsecase interface {
	CreateGame(ctx context.Context, req httpModel.CreateGameRequest) error
}

func (d *delivery) CreateGame(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req httpModel.CreateGameRequest
	err := UnmarshalRequest(r.Body, &req)
	if err != nil {
		ResponseWithJSON(w, http.StatusBadRequest, Error{Msg: err.Error()})
		return
	}

	err = d.usecase.CreateGame(ctx, req)
	if err != nil {
		ResponseWithJSON(w, http.StatusInternalServerError, Error{Msg: err.Error()})
		return
	}

	ResponseWithJSON(w, http.StatusOK, nil)
}
