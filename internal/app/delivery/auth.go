package delivery

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"quizon/internal/app/usecase"
)

// sessionCookie - кука с токеном сессии админа
const sessionCookie string = `authorization-token`

func (wb *delivery) isAdmin(r *http.Request) bool {
	cookie, err := r.Cookie(sessionCookie)
	if err != nil {
		return false
	}

	return wb.usecase.CheckSession(cookie.Value)
}

func (wb *delivery) requireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !wb.isAdmin(r) {
			redirect(w, r, "/login")
			return
		}

		next.ServeHTTP(w, r)
	})
}

type loginView struct {
	Login string
	Error string
}

func (wb *delivery) loginPage(w http.ResponseWriter, r *http.Request) {
	if wb.isAdmin(r) {
		redirect(w, r, "/")
		return
	}

	wb.renderPage(w, r, http.StatusOK, pageLogin, loginView{})
}

func (wb *delivery) login(w http.ResponseWriter, r *http.Request) {
	view := loginView{
		Login: strings.TrimSpace(r.PostFormValue("login")),
	}
	password := r.PostFormValue("password")

	if view.Login == "" || password == "" {
		view.Error = "Введите логин и пароль"
		wb.renderPage(w, r, http.StatusUnprocessableEntity, pageLogin, view)
		return
	}

	session, err := wb.usecase.Login(r.Context(), view.Login, password)
	if errors.Is(err, usecase.ErrWrongPassword) {
		view.Error = "Неверный логин или пароль"
		wb.renderPage(w, r, http.StatusUnprocessableEntity, pageLogin, view)
		return
	}
	if err != nil {
		wb.serverError(w, r, err)
		return
	}

	http.SetCookie(w, wb.newSessionCookie(session.Token, session.ExpiresAt))
	redirect(w, r, "/")
}

func (wb *delivery) logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(sessionCookie)
	if err == nil {
		wb.usecase.Logout(cookie.Value)
	}

	// браузер удалит куку, только если имя, путь и домен совпадают с кукой входа
	http.SetCookie(w, wb.newSessionCookie("", time.Unix(0, 0)))
	redirect(w, r, "/")
}

// newSessionCookie - кука сессии админа: только по https, из js не видна
func (wb *delivery) newSessionCookie(token string, expires time.Time) *http.Cookie {
	return &http.Cookie{
		Name:     sessionCookie,
		Value:    token,
		Expires:  expires,
		Path:     "/",
		Domain:   wb.domain,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
}
