package delivery

import (
	"context"
	"net/http"

	"github.com/rs/zerolog/log"
)

type CheckCookieUsecase interface {
	CheckCookie(ctx context.Context, cookie string) (bool, error)
}

func (d delivery) CheckAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("auth")
		if err != nil {
			log.Info().Err(err).Msg("can't get auth cookie")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		ok, err := d.checkAuthUsecase.CheckCookie(r.Context(), cookie.Value)
		if err != nil {
			log.Error().Err(err).Msg("")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
