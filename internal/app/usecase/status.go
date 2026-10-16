package usecase

// RegistrationStatus - куда попадает следующая команда
type RegistrationStatus string

const (
	// StatusOk - в основной состав
	StatusOk RegistrationStatus = "ok"
	// StatusReserve - в резерв
	StatusReserve RegistrationStatus = "reserve"
	// StatusClosed - мест нет
	StatusClosed RegistrationStatus = "closed"
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
