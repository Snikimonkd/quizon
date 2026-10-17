package web

import (
	"errors"
	"net/http"
	"strings"

	"quizon/internal/app/usecase"
)

// sessionCookie - кука с токеном сессии админа
const sessionCookie string = `authorization-token`

func (wb *Web) isAdmin(r *http.Request) bool {
	cookie, err := r.Cookie(sessionCookie)
	if err != nil {
		return false
	}

	return wb.usecase.CheckSession(cookie.Value)
}

func (wb *Web) requireAdmin(next http.Handler) http.Handler {
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

func (wb *Web) loginPage(w http.ResponseWriter, r *http.Request) {
	if wb.isAdmin(r) {
		redirect(w, r, "/")
		return
	}

	wb.renderPage(w, r, http.StatusOK, pageLogin, loginView{})
}

func (wb *Web) login(w http.ResponseWriter, r *http.Request) {
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

	http.SetCookie(w, &http.Cookie{
		Name:    sessionCookie,
		Value:   session.Token,
		Expires: session.ExpiresAt,
		Path:    "/",
		Domain:  wb.domain,
		// только по https, из js не видна
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	redirect(w, r, "/")
}

func (wb *Web) logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(sessionCookie)
	if err == nil {
		wb.usecase.Logout(cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	redirect(w, r, "/")
}
