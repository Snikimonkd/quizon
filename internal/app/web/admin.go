package web

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/samber/lo"

	"quizon/internal/app/usecase"
	"quizon/internal/generated/postgres/public/model"
	"quizon/internal/utils"
)

// datetimeLocal - формат значения <input type="datetime-local">
const datetimeLocal string = "2006-01-02T15:04"

const (
	defaultGameName      string = "Бауманская лига"
	defaultGameLocation  string = "МГТУ им. Н.Э.Баумана ауд. 345"
	defaultMainAmount    string = "20"
	defaultReserveAmount string = "10"
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

// validate - ошибки разбора формы: даты в формате datetime-local, количества — целые числа.
// Пустое число не разобрать, поэтому его наличие проверяется здесь. Остальные правила игры проверяет usecase
func (f gameForm) validate() map[string]string {
	return fieldErrors(validation.ValidateStruct(
		&f,
		validation.Field(&f.StartTime, moscowTime),
		validation.Field(&f.RegistrationOpenTime, moscowTime),
		validation.Field(
			&f.MainAmount,
			validation.Required.Error("Сколько команд играет?"),
			integer("Нужно число от 1"),
		),
		validation.Field(&f.ReserveAmount, integer("Нужно число от 0")),
	))
}

// gameRuleMessages - подсказки у полей формы игры на нарушенные правила usecase
var gameRuleMessages = ruleMessages{
	usecase.FieldName:      {usecase.ErrRequired: "Укажите название"},
	usecase.FieldLocation:  {usecase.ErrRequired: "Укажите место"},
	usecase.FieldStartTime: {usecase.ErrRequired: "Укажите дату и время игры"},
	usecase.FieldRegistrationOpenTime: {
		usecase.ErrRequired:        "Укажите, когда открыть регистрацию",
		usecase.ErrOpensAfterStart: "Регистрация должна открыться до начала игры",
	},
	usecase.FieldMainAmount:    {usecase.ErrTooSmall: "Нужно число от 1"},
	usecase.FieldReserveAmount: {usecase.ErrTooSmall: "Нужно число от 0"},
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
			Name: "registration_open_time", Label: "Открытие регистрации", Type: inputDateTime,
			Value: f.RegistrationOpenTime, Hint: "По Москве", Required: true, Error: errs["RegistrationOpenTime"],
		},
		{
			Name: "location", Label: "Место проведения", Type: inputText, Value: f.Location,
			Required: true, Error: errs["Location"],
		},
		{
			Name: "main_amount", Label: "Команд в основном составе", Type: inputNumber, Value: f.MainAmount,
			InputMode: inputModeNumeric, Min: "1", Required: true, Error: errs["MainAmount"],
		},
		{
			Name: "reserve_amount", Label: "Команд в резерве", Type: inputNumber, Value: f.ReserveAmount,
			InputMode: inputModeNumeric, Min: "0", Error: errs["ReserveAmount"],
		},
	}
}

type gameFormView struct {
	Fields []field
}

func (wb *Web) newGamePage(w http.ResponseWriter, r *http.Request) {
	form := gameForm{
		Name:          defaultGameName,
		Location:      defaultGameLocation,
		MainAmount:    defaultMainAmount,
		ReserveAmount: defaultReserveAmount,
	}
	wb.renderPage(w, r, http.StatusOK, pageGameNew, gameFormView{Fields: form.fields(nil)})
}

func (wb *Web) createGame(w http.ResponseWriter, r *http.Request) {
	form := parseGameForm(r)

	errs := form.validate()
	if len(errs) > 0 {
		wb.renderGameForm(w, r, form, errs)
		return
	}

	// правила игры проверяет usecase, здесь только показываем нарушения у полей
	_, err := wb.usecase.CreateGame(r.Context(), form.game())
	invalid, isInvalid := errors.AsType[usecase.ValidationError](err)
	if isInvalid {
		wb.renderGameForm(w, r, form, gameRuleMessages.messages(invalid))
		return
	}
	if err != nil {
		wb.serverError(w, r, err)
		return
	}

	redirect(w, r, "/")
}

// renderGameForm - форма новой игры с ошибками errs у полей
func (wb *Web) renderGameForm(w http.ResponseWriter, r *http.Request, form gameForm, errs map[string]string) {
	wb.renderPage(w, r, http.StatusUnprocessableEntity, pageGameNew, gameFormView{Fields: form.fields(errs)})
}

type teamsView struct {
	Game gameView
	// Teams - все регистрации в порядке подачи, удалённые тоже: в таблице они серые и без номера
	Teams []usecase.Registration
	// ActiveCount - сколько команд осталось в игре, без удалённых
	ActiveCount int
	// MainAmount - мест в основе, как задали при создании игры
	MainAmount int64
	// ReserveAmount - мест в резерве, как задали при создании игры
	ReserveAmount int64
}

// FirstReserveName - название первой команды резерва: она займёт место команды, удалённой из основы.
// Пусто, если резерва нет
func (v teamsView) FirstReserveName() string {
	team, _ := lo.Find(v.Teams, func(t usecase.Registration) bool { return t.FirstReserve })

	return team.TeamName
}

func (wb *Web) listTeams(w http.ResponseWriter, r *http.Request) (teamsView, bool) {
	game, ok := wb.loadGame(w, r)
	if !ok {
		return teamsView{}, false
	}

	teams, err := wb.usecase.ListRegistrations(r.Context(), game.ID)
	if err != nil {
		wb.serverError(w, r, err)
		return teamsView{}, false
	}

	return teamsView{
		Game:          newGameView(game, time.Now()),
		Teams:         teams,
		ActiveCount:   lo.CountBy(teams, func(t usecase.Registration) bool { return !t.Deleted }),
		MainAmount:    game.MainAmount,
		ReserveAmount: game.ReserveAmount,
	}, true
}

func (wb *Web) teamsPage(w http.ResponseWriter, r *http.Request) {
	view, ok := wb.listTeams(w, r)
	if !ok {
		return
	}

	wb.renderPage(w, r, http.StatusOK, pageTeams, view)
}

// deleteRegistration - убирает команду из игры и показывает обновлённый список: номера и граница резерва сдвигаются.
// В адресе id регистрации (registrations.id), не «ID команды» из анкеты
func (wb *Web) deleteRegistration(w http.ResponseWriter, r *http.Request) {
	gameID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	registrationID, err := strconv.ParseInt(chi.URLParam(r, "registrationID"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// ErrNotFound - команду уже удалили, например двойным нажатием: просто показываем актуальный список
	err = wb.usecase.DeleteRegistration(r.Context(), gameID, registrationID)
	if err != nil && !errors.Is(err, usecase.ErrNotFound) {
		wb.serverError(w, r, err)
		return
	}

	if !isHTMX(r) {
		redirect(w, r, fmt.Sprintf("/admin/games/%d/teams", gameID))
		return
	}

	view, ok := wb.listTeams(w, r)
	if !ok {
		return
	}

	wb.renderBlock(w, r, http.StatusOK, pageTeams, "teams", view)
}
