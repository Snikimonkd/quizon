package web

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"quizon/internal/app/usecase"
	"quizon/internal/generated/postgres/public/model"
	"quizon/internal/utils"
)

// datetimeLocal - формат значения <input type="datetime-local">
const datetimeLocal string = "2006-01-02T15:04"

const (
	defaultGameName     string = "Бауманская лига"
	defaultGameLocation string = "МГТУ им. Н.Э.Баумана ауд. 345"
)

type gameForm struct {
	Name                 string
	Location             string
	StartTime            string
	RegistrationOpenTime string
	MainAmount           string
	ReserveAmount        string
}

func parseGameForm(r *http.Request) gameForm {
	value := func(name string) string {
		return strings.TrimSpace(r.PostFormValue(name))
	}

	return gameForm{
		Name:                 value("name"),
		Location:             value("location"),
		StartTime:            value("start_time"),
		RegistrationOpenTime: value("registration_open_time"),
		MainAmount:           value("main_amount"),
		ReserveAmount:        value("reserve_amount"),
	}
}

// moscowTime - правило: строка из datetime-local, время московское
var moscowTime = validation.By(func(value any) error {
	s, _ := value.(string)
	if s == "" {
		return nil
	}

	_, err := time.ParseInLocation(datetimeLocal, s, utils.LocMsk)
	if err != nil {
		return validation.NewError("datetime", "Укажите дату и время")
	}

	return nil
})

func (f gameForm) validate() map[string]string {
	errs := fieldErrors(validation.ValidateStruct(
		&f,
		validation.Field(&f.Name, validation.Required.Error("Укажите название")),
		validation.Field(&f.Location, validation.Required.Error("Укажите место")),
		validation.Field(&f.StartTime, validation.Required.Error("Укажите дату и время игры"), moscowTime),
		validation.Field(
			&f.RegistrationOpenTime,
			validation.Required.Error("Укажите, когда открыть регистрацию"),
			moscowTime,
		),
		validation.Field(
			&f.MainAmount,
			validation.Required.Error("Сколько команд играет?"),
			intAtLeast(1, "Нужно число от 1"),
		),
		validation.Field(&f.ReserveAmount, intAtLeast(0, "Нужно число от 0")),
	))
	if len(errs) > 0 {
		return errs
	}

	start, open := f.times()
	if !open.Before(start) {
		return map[string]string{"RegistrationOpenTime": "Регистрация должна открыться до начала игры"}
	}

	return nil
}

func (f gameForm) times() (time.Time, time.Time) {
	start, _ := time.ParseInLocation(datetimeLocal, f.StartTime, utils.LocMsk)
	open, _ := time.ParseInLocation(datetimeLocal, f.RegistrationOpenTime, utils.LocMsk)

	return start, open
}

func (f gameForm) game() model.Games {
	start, open := f.times()
	mainAmount, _ := strconv.ParseInt(f.MainAmount, 10, 64)
	reserveAmount, _ := strconv.ParseInt(f.ReserveAmount, 10, 64)

	return model.Games{
		Name:                 f.Name,
		Location:             f.Location,
		StartTime:            start,
		RegistrationOpenTime: open,
		MainAmount:           mainAmount,
		ReserveAmount:        reserveAmount,
	}
}

func (f gameForm) fields(errs map[string]string) []field {
	return []field{
		{
			Name: "name", Label: "Название", Type: inputText, Value: f.Name,
			Required: true, Error: errs["Name"],
		},
		{
			Name: "start_time", Label: "Дата и время игры", Type: inputDateTime, Value: f.StartTime,
			Hint: "По Москве", Required: true, Error: errs["StartTime"],
		},
		{
			Name: "location", Label: "Место проведения", Type: inputText, Value: f.Location,
			Required: true, Error: errs["Location"],
		},
		{
			Name: "main_amount", Label: "Команд в основном составе", Type: inputNumber, Value: f.MainAmount,
			InputMode: inputModeNumeric, Min: "1", Placeholder: "30", Required: true, Error: errs["MainAmount"],
		},
		{
			Name: "reserve_amount", Label: "Команд в резерве", Type: inputNumber, Value: f.ReserveAmount,
			InputMode: inputModeNumeric, Min: "0", Placeholder: "5", Error: errs["ReserveAmount"],
		},
		{
			Name: "registration_open_time", Label: "Открытие регистрации", Type: inputDateTime,
			Value: f.RegistrationOpenTime, Hint: "По Москве", Required: true, Error: errs["RegistrationOpenTime"],
		},
	}
}

type gameFormView struct {
	Fields []field
}

func (wb *Web) newGamePage(w http.ResponseWriter, r *http.Request) {
	form := gameForm{Name: defaultGameName, Location: defaultGameLocation, ReserveAmount: "0"}
	wb.renderPage(w, http.StatusOK, "game_new", wb.layout(r, "Новая игра", gameFormView{Fields: form.fields(nil)}))
}

func (wb *Web) createGame(w http.ResponseWriter, r *http.Request) {
	form := parseGameForm(r)

	errs := form.validate()
	if len(errs) > 0 {
		view := gameFormView{Fields: form.fields(errs)}
		wb.renderPage(w, http.StatusUnprocessableEntity, "game_new", wb.layout(r, "Новая игра", view))
		return
	}

	_, err := wb.usecase.CreateGame(r.Context(), form.game())
	if err != nil {
		wb.serverError(w, err)
		return
	}

	redirect(w, r, "/")
}

type teamsView struct {
	Game  gameView
	Teams []usecase.Registration
	// MainAmount - команды с номером больше этого попали в резерв
	MainAmount int64
}

func (wb *Web) listTeams(w http.ResponseWriter, r *http.Request) (teamsView, bool) {
	game, ok := wb.loadGame(w, r)
	if !ok {
		return teamsView{}, false
	}

	teams, err := wb.usecase.ListRegistrations(r.Context(), game.ID)
	if err != nil {
		wb.serverError(w, err)
		return teamsView{}, false
	}

	return teamsView{
		Game:       newGameView(game, usecase.StatusOk, wb.now()),
		Teams:      teams,
		MainAmount: game.MainAmount,
	}, true
}

func (wb *Web) teamsPage(w http.ResponseWriter, r *http.Request) {
	view, ok := wb.listTeams(w, r)
	if !ok {
		return
	}

	wb.renderPage(w, http.StatusOK, "teams", wb.layout(r, "Участники", view))
}
