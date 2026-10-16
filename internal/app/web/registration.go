package web

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"unicode"

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

func (f registrationForm) validate() map[string]string {
	return fieldErrors(validation.ValidateStruct(
		&f,
		validation.Field(&f.CaptainName, validation.Required.Error("Укажите имя капитана")),
		validation.Field(&f.Phone, validation.Required.Error("Укажите номер телефона"), validation.By(phoneRule)),
		validation.Field(&f.Telegram, validation.Required.Error("Укажите ник в Telegram")),
		validation.Field(&f.TeamName, validation.Required.Error("Укажите название команды")),
		validation.Field(
			&f.PlayersAmount,
			validation.Required.Error("Укажите, сколько человек придёт"),
			intAtLeast(1, "Нужно число от 1"),
		),
	))
}

// phoneRule - в номере должно быть от 10 до 15 цифр, остальное (+, скобки, пробелы, дефисы) не важно
func phoneRule(value any) error {
	s, _ := value.(string)
	digits := 0
	for _, r := range s {
		if unicode.IsDigit(r) {
			digits++
		}
	}

	if digits < 10 || digits > 15 {
		return validation.NewError("phone", "Похоже, в номере ошибка")
	}

	return nil
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
		Game:   newGameView(game, usecase.StatusOk, wb.now()),
		Fields: registrationForm{}.fields(nil),
	}
	view.Outcome = closedOutcome(view.Game)
	wb.renderPage(w, http.StatusOK, "register", wb.layout(r, "Регистрация на игру", view))
}

func (wb *Web) register(w http.ResponseWriter, r *http.Request) {
	game, ok := wb.loadGame(w, r)
	if !ok {
		return
	}

	form := parseRegistrationForm(r)
	view := registerView{Game: newGameView(game, usecase.StatusOk, wb.now())}

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
	switch {
	case errors.Is(err, usecase.ErrNotOpenedYet):
		view.Outcome = outcomeWaiting
	case err != nil:
		wb.serverError(w, err)
		return
	default:
		view.Outcome = outcome(status)
	}

	wb.renderRegistration(w, r, http.StatusOK, view)
}

// closedOutcome - почему форму сейчас показывать не надо; пустая строка - форма нужна
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
		wb.renderBlock(w, status, "register", "registration", view)
		return
	}

	wb.renderPage(w, status, "register", wb.layout(r, "Регистрация на игру", view))
}
