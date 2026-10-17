package web

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"quizon/internal/app/usecase"
	"quizon/internal/generated/postgres/public/model"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/samber/lo"
)

// outcome - чем закончилась попытка зарегистрироваться:
// один из usecase.RegistrationStatus (ok, reserve, closed) или waiting/finished
type outcome string

const (
	outcomeWaiting  outcome = "waiting"
	outcomeFinished outcome = "finished"
)

type registrationForm struct {
	CaptainName   string
	GroupName     string
	Phone         string
	Telegram      string
	TeamName      string
	TeamID        string
	PlayersAmount string
}

func parseRegistrationForm(r *http.Request) registrationForm {
	value := func(name string) string {
		return strings.TrimSpace(r.PostFormValue(name))
	}

	return registrationForm{
		CaptainName:   value("captain_name"),
		GroupName:     value("group_name"),
		Phone:         value("phone"),
		Telegram:      value("telegram"),
		TeamName:      value("team_name"),
		TeamID:        value("team_id"),
		PlayersAmount: value("players_amount"),
	}
}

// validate - ошибки разбора анкеты: сколько человек придёт — целое число. Пустое число не разобрать,
// поэтому его наличие проверяется здесь. Остальные правила анкеты проверяет usecase
func (f registrationForm) validate() map[string]string {
	return fieldErrors(validation.ValidateStruct(
		&f,
		validation.Field(
			&f.PlayersAmount,
			validation.Required.Error("Укажите, сколько человек придёт"),
			integer("Нужно число от 1"),
		),
	))
}

// registrationRuleMessages - подсказки у полей анкеты на нарушенные правила usecase
var registrationRuleMessages = ruleMessages{
	usecase.FieldCaptainName: {usecase.ErrRequired: "Укажите имя капитана"},
	usecase.FieldPhone: {
		usecase.ErrRequired: "Укажите номер телефона",
		usecase.ErrBadPhone: "Похоже, в номере ошибка",
	},
	usecase.FieldTelegram:      {usecase.ErrRequired: "Укажите ник в Telegram"},
	usecase.FieldTeamName:      {usecase.ErrRequired: "Укажите название команды"},
	usecase.FieldPlayersAmount: {usecase.ErrTooSmall: "Нужно число от 1"},
}

func (f registrationForm) registration(gameID int64) model.Registrations {
	amount, _ := strconv.ParseInt(f.PlayersAmount, 10, 64)

	return model.Registrations{
		GameID:        gameID,
		CaptainName:   f.CaptainName,
		GroupName:     lo.EmptyableToPtr(f.GroupName),
		Phone:         f.Phone,
		PlayersAmount: amount,
		TeamID:        lo.EmptyableToPtr(f.TeamID),
		TeamName:      f.TeamName,
		Telegram:      f.Telegram,
	}
}

func (f registrationForm) fields(errs map[string]string) []field {
	return []field{
		{
			Name: "captain_name", Label: "Имя капитана", Type: inputText, Value: f.CaptainName,
			Placeholder: "Николай Эрнестович Бауман", Autocomplete: "name", Required: true,
			Error: errs["CaptainName"],
		},
		{
			Name: "group_name", Label: "Учебная группа капитана", Type: inputText, Value: f.GroupName,
			Placeholder: "СМ1-11", Error: errs["GroupName"],
		},
		{
			Name: "phone", Label: "Номер телефона", Type: inputTel, Value: f.Phone,
			Placeholder: "8 999 888-77-66", Autocomplete: inputTel, InputMode: inputTel, Required: true,
			Error: errs["Phone"],
		},
		{
			Name: "telegram", Label: "Telegram", Type: inputText, Value: f.Telegram,
			Placeholder: "@quizonmsk", Autocomplete: autocompleteOff, Required: true, Error: errs["Telegram"],
		},
		{
			Name: "team_name", Label: "Название команды", Type: inputText, Value: f.TeamName,
			Placeholder: "КвизON", Autocomplete: autocompleteOff, Required: true, Error: errs["TeamName"],
		},
		{
			Name: "team_id", Label: "ID команды", Type: inputText, Value: f.TeamID,
			Placeholder: "123456", Autocomplete: autocompleteOff, Hint: "Если команда уже играла в лиге",
			Error: errs["TeamID"],
		},
		{
			Name: "players_amount", Label: "Сколько человек придёт", Type: inputNumber, Value: f.PlayersAmount,
			Placeholder: "6", InputMode: inputModeNumeric, Min: "1", Required: true, Error: errs["PlayersAmount"],
		},
	}
}

type registerView struct {
	Game    gameView
	Fields  []field
	Outcome outcome
}

func (wb *Web) registrationPage(w http.ResponseWriter, r *http.Request) {
	game, ok := wb.loadGame(w, r)
	if !ok {
		return
	}

	view := registerView{
		Game:   newGameView(game, time.Now()),
		Fields: registrationForm{}.fields(nil),
	}
	view.Outcome = closedOutcome(view.Game)
	// мест нет - вместо формы сразу итог
	if view.Game.State == stateClosed {
		view.Outcome = outcome(usecase.StatusClosed)
	}
	wb.renderPage(w, r, http.StatusOK, pageRegister, view)
}

func (wb *Web) register(w http.ResponseWriter, r *http.Request) {
	game, ok := wb.loadGame(w, r)
	if !ok {
		return
	}

	form := parseRegistrationForm(r)
	view := registerView{Game: newGameView(game, time.Now())}

	view.Outcome = closedOutcome(view.Game)
	if view.Outcome != "" {
		wb.renderRegistration(w, r, http.StatusOK, view)
		return
	}

	errs := form.validate()
	if len(errs) > 0 {
		view.Fields = form.fields(errs)
		wb.renderRegistration(w, r, http.StatusUnprocessableEntity, view)
		return
	}

	status, err := wb.usecase.CreateRegistration(r.Context(), form.registration(game.ID))
	invalid, isInvalid := errors.AsType[usecase.ValidationError](err)
	switch {
	case isInvalid:
		// правила анкеты проверяет usecase, здесь только показываем нарушения у полей
		view.Fields = form.fields(registrationRuleMessages.messages(invalid))
		wb.renderRegistration(w, r, http.StatusUnprocessableEntity, view)
		return
	case errors.Is(err, usecase.ErrNotOpenedYet):
		view.Outcome = outcomeWaiting
	case err != nil:
		wb.serverError(w, r, err)
		return
	default:
		view.Outcome = outcome(status)
	}

	wb.renderRegistration(w, r, http.StatusOK, view)
}

// closedOutcome - итог вместо формы, если регистрация закрыта по времени: ещё не открылась или игра прошла.
// Пустая строка - форма нужна. Места здесь не проверяем: при отправке их считает usecase под блокировкой игры
func closedOutcome(game gameView) outcome {
	switch game.State {
	case stateFinished:
		return outcomeFinished
	case stateWaiting:
		return outcomeWaiting
	case stateOpen, stateReserve, stateClosed:
	}

	return ""
}

func (wb *Web) renderRegistration(w http.ResponseWriter, r *http.Request, status int, view registerView) {
	if isHTMX(r) {
		wb.renderBlock(w, r, status, pageRegister, "registration", view)
		return
	}

	wb.renderPage(w, r, status, pageRegister, view)
}
