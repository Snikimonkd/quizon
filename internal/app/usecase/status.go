package usecase

import (
	"time"

	"quizon/internal/generated/postgres/public/model"
)

// RegistrationStatus - что сейчас с регистрацией на игру и куда попадает следующая команда
type RegistrationStatus string

const (
	// StatusOk - в основной состав
	StatusOk RegistrationStatus = "ok"
	// StatusReserve - в резерв
	StatusReserve RegistrationStatus = "reserve"
	// StatusClosed - мест нет
	StatusClosed RegistrationStatus = "closed"
	// StatusWaiting - регистрация ещё не открылась
	StatusWaiting RegistrationStatus = "waiting"
	// StatusFinished - игра уже началась или прошла
	StatusFinished RegistrationStatus = "finished"
)

// registrationStatus - статус для следующей команды, если уже зарегистрировано registered
func registrationStatus(registered, mainAmount, reserveAmount int64) RegistrationStatus {
	switch {
	case registered >= mainAmount+reserveAmount:
		return StatusClosed
	case registered >= mainAmount:
		return StatusReserve
	default:
		return StatusOk
	}
}

// gameStatus - статус регистрации на игру в момент now: время важнее мест.
// Игра началась — регистрации больше нет, время открытия не пришло — ещё нет, иначе решают места
func gameStatus(game model.Games, registered int64, now time.Time) RegistrationStatus {
	switch {
	case !now.Before(game.StartTime):
		return StatusFinished
	case now.Before(game.RegistrationOpenTime):
		return StatusWaiting
	default:
		return registrationStatus(registered, game.MainAmount, game.ReserveAmount)
	}
}
