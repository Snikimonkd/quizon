package testsupport

import (
	"github.com/brianvoe/gofakeit/v7"

	"quizon/internal/generated/postgres/public/model"
)

// FakeRegistration - регистрация команды на игру со случайными данными капитана и команды.
// Время подачи не заполняет: порядок регистраций каждый тест задаёт сам
func FakeRegistration(gameID int64) model.Registrations {
	return model.Registrations{
		GameID:      gameID,
		TeamName:    gofakeit.Gamertag(),
		CaptainName: gofakeit.Name(),
		// номер с разделителями, как его вводят в анкете: (974)536-0611, 1-769-160-8766
		Phone:         gofakeit.PhoneFormatted(),
		Telegram:      "@" + gofakeit.Username(),
		PlayersAmount: int64(gofakeit.IntRange(1, 10)),
		// учебная группа МГТУ, например ИУ7-54Б
		GroupName: new(gofakeit.Regex(`(ИУ|РК|СМ|МТ|РЛ|БМТ|Э)[1-9]{1,2}-[1-9][1-9]Б?`)),
		TeamID:    new(gofakeit.DigitN(5)),
	}
}
