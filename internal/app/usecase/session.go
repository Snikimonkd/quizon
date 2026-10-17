package usecase

// CheckSession - действует ли сессия админа с этим токеном. Истёкшую сессию сразу удаляет
func (u usecase) CheckSession(token string) bool {
	cookie, ok := u.cookieCache.Get(token)
	if !ok {
		return false
	}

	if cookie.Expires.Before(u.clock.Now()) {
		u.cookieCache.Del(token)
		return false
	}

	return true
}

// Logout - завершает сессию админа
func (u usecase) Logout(token string) {
	u.cookieCache.Del(token)
}
