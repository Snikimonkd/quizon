package web

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"quizon/internal/app/usecase"

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
	// RegistrationOpens - когда откроется регистрация словами, см. ruOpensAt
	RegistrationOpens string
	State             gameState
	// Admin - смотрит организатор: вместо регистрации показываем список участников
	Admin bool
}

// newGameView - игра для шаблонов; состояние считается из времени и статуса регистрации
func newGameView(game usecase.GameListItem, now time.Time) gameView {
	view := gameView{
		ID:                   game.ID,
		Name:                 game.Name,
		Location:             game.Location,
		StartTime:            game.StartTime,
		RegistrationOpenTime: game.RegistrationOpenTime,
		RegistrationOpens:    ruOpensAt(game.RegistrationOpenTime, now),
		State:                stateOpen,
	}

	switch {
	case !now.Before(game.StartTime):
		view.State = stateFinished
	case now.Before(game.RegistrationOpenTime):
		view.State = stateWaiting
	case game.Status == usecase.StatusReserve:
		view.State = stateReserve
	case game.Status == usecase.StatusClosed:
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
		wb.serverError(w, r, err)
		return
	}

	now := wb.now()
	admin := wb.isAdmin(r)
	view := homeView{Games: make([]gameView, 0, len(items))}
	for _, item := range items {
		gv := newGameView(item, now)
		gv.Admin = admin
		view.Games = append(view.Games, gv)
	}

	wb.renderPage(w, r, http.StatusOK, pageHome, view)
}

// loadGame - достаёт игру из {id} в пути вместе со статусом регистрации; при ошибке сам пишет ответ
func (wb *Web) loadGame(w http.ResponseWriter, r *http.Request) (usecase.GameListItem, bool) {
	gameID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return usecase.GameListItem{}, false
	}

	game, err := wb.usecase.GetGame(r.Context(), gameID)
	if errors.Is(err, usecase.ErrNotFound) {
		http.NotFound(w, r)
		return usecase.GameListItem{}, false
	}
	if err != nil {
		wb.serverError(w, r, err)
		return usecase.GameListItem{}, false
	}

	return game, true
}
