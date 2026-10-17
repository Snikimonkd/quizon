package usecase

import "time"

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

// Now - текущее время по часам usecase. Веб своё время не берёт:
// состояние игр считает usecase, а подписи вроде «завтра в 18:00» и год в подвале — от этих же часов
func (u usecase) Now() time.Time {
	return u.clock.Now()
}
