package usecase

import "time"

// Session - сессия админа: находится по токену и действует до ExpiresAt
type Session struct {
	Token     string
	ExpiresAt time.Time
}

// SessionStore - хранилище сессий админов по токену
type SessionStore interface {
	Set(token string, session Session)
	Get(token string) (Session, bool)
	Del(token string)
}

// CheckSession - действует ли сессия админа с этим токеном. Истёкшую сессию сразу удаляет
func (u usecase) CheckSession(token string) bool {
	session, ok := u.sessions.Get(token)
	if !ok {
		return false
	}

	if session.ExpiresAt.Before(u.clock.Now()) {
		u.sessions.Del(token)
		return false
	}

	return true
}

// Logout - завершает сессию админа
func (u usecase) Logout(token string) {
	u.sessions.Del(token)
}
