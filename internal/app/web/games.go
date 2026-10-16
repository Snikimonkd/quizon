package web

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"quizon/internal/app/usecase"
	"quizon/internal/generated/postgres/public/model"

	"github.com/go-chi/chi/v5"
)

// gamesPerPage - сколько игр показываем на главной
const gamesPerPage int64 = 100

// gameState - что сейчас можно сделать с игрой
type gameState string

const (
	stateWaiting  gameState = "waiting"
	stateOpen     gameState = "open"
	stateReserve  gameState = "reserve"
	stateClosed   gameState = "closed"
	stateFinished gameState = "finished"
)

type gameView struct {
	ID                   int64
	Name                 string
	Location             string
	StartTime            time.Time
	RegistrationOpenTime time.Time
	State                gameState
	// Admin - смотрит организатор: вместо регистрации показываем список участников
	Admin bool
}

func newGameView(game model.Games, status usecase.RegistrationStatus, now time.Time) gameView {
	view := gameView{
		ID:                   game.ID,
		Name:                 game.Name,
		Location:             game.Location,
		StartTime:            game.StartTime,
		RegistrationOpenTime: game.RegistrationOpenTime,
		State:                stateOpen,
	}

	switch {
	case !now.Before(game.StartTime):
		view.State = stateFinished
	case now.Before(game.RegistrationOpenTime):
		view.State = stateWaiting
	case status == usecase.StatusReserve:
		view.State = stateReserve
	case status == usecase.StatusClosed:
		view.State = stateClosed
	}

	return view
}

type homeView struct {
	Games []gameView
}

func (wb *Web) home(w http.ResponseWriter, r *http.Request) {
	items, err := wb.usecase.ListGames(r.Context(), 1, gamesPerPage)
	if err != nil {
		wb.serverError(w, err)
		return
	}

	now := wb.now()
	admin := wb.isAdmin(r)
	view := homeView{Games: make([]gameView, 0, len(items))}
	for _, item := range items {
		gv := newGameView(item.Games, item.Status, now)
		gv.Admin = admin
		view.Games = append(view.Games, gv)
	}

	wb.renderPage(w, r, http.StatusOK, pageHome, view)
}

// loadGame - достаёт игру из {id} в пути; при ошибке сам пишет ответ
func (wb *Web) loadGame(w http.ResponseWriter, r *http.Request) (model.Games, bool) {
	gameID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return model.Games{}, false
	}

	game, err := wb.usecase.GetGame(r.Context(), gameID)
	if errors.Is(err, usecase.ErrNotFound) {
		http.NotFound(w, r)
		return model.Games{}, false
	}
	if err != nil {
		wb.serverError(w, err)
		return model.Games{}, false
	}

	return game, true
}
