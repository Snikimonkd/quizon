package http

import (
	"context"
	"net/http"
	"time"

	"quizon/internal/app/usecase"
	"quizon/internal/pkg/logger"
)

type loginKey string
type loginValue string

type CookieCache interface {
	Get(key string) (usecase.Cookie, bool)
	Del(key string)
}

type CheckCookieMiddleware struct {
	cache CookieCache
}

func NewCheckCookieMiddleware(cache CookieCache) CheckCookieMiddleware {
	return CheckCookieMiddleware{
		cache: cache,
	}
}

func (cm CheckCookieMiddleware) CheckCookie() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(authorizationTokenName)
			if err != nil {
				logger.Warnf("no cookie in request")
				ResponseWithJSON(w, http.StatusUnauthorized, nil)
				return
			}

			knownCookie, ok := cm.cache.Get(cookie.Value)
			if !ok {
				logger.Warnf("unknown cookie")
				ResponseWithJSON(w, http.StatusUnauthorized, nil)
				return
			}

			if knownCookie.ExpiresAt.Before(time.Now()) {
				logger.Warnf("cookie is expierd")
				cm.cache.Del(cookie.Value)
				ResponseWithJSON(w, http.StatusUnauthorized, nil)
				return
			}

			ctx := r.Context()
			ctx = context.WithValue(ctx, loginKey("login"), loginValue(knownCookie.Login))
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}
