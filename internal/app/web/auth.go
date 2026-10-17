package web

import (
	"errors"
	"net/http"
	"strings"

	"quizon/internal/app/usecase"
)

func (wb *Web) isAdmin(r *http.Request) bool {
	cookie, err := r.Cookie(usecase.AuthorizationTokenName)
	if err != nil {
		return false
	}

	known, ok := wb.cache.Get(cookie.Value)
	if !ok {
		return false
	}

	if known.Expires.Before(wb.now()) {
		wb.cache.Del(cookie.Value)
		return false
	}

	return true
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

	cookie, err := wb.usecase.Login(r.Context(), view.Login, password)
	if errors.Is(err, usecase.ErrWrongPassword) {
		view.Error = "Неверный логин или пароль"
		wb.renderPage(w, r, http.StatusUnprocessableEntity, pageLogin, view)
		return
	}
	if err != nil {
		wb.serverError(w, r, err)
		return
	}

	http.SetCookie(w, cookie)
	redirect(w, r, "/")
}

func (wb *Web) logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(usecase.AuthorizationTokenName)
	if err == nil {
		wb.cache.Del(cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     usecase.AuthorizationTokenName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	redirect(w, r, "/")
}
