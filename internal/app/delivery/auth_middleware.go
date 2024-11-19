package delivery

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"quizon/internal/app/delivery/api"

	strictnethttp "github.com/oapi-codegen/runtime/strictmiddleware/nethttp"
	"github.com/samber/lo"
)

const AuthorizationTokenName string = `authorization-token`

// type loginKey string
// type loginValue string

type CookieCache interface {
	Get(key string) (*http.Cookie, bool)
	Del(key string)
}

type checkCookieMiddleware struct {
	cache CookieCache
}

func NewCheckCookieMiddleware(cache CookieCache) checkCookieMiddleware {
	return checkCookieMiddleware{
		cache: cache,
	}
}

type unauthorizedResponse struct {
	Body api.Error
}

func (u unauthorizedResponse) VisitPostGameResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)

	return json.NewEncoder(w).Encode(u.Body)
}

func (u unauthorizedResponse) VisitGetAuthResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)

	return json.NewEncoder(w).Encode(u.Body)
}

func (u unauthorizedResponse) VisitGetGamesIdRegistrationsResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)

	return json.NewEncoder(w).Encode(u.Body)
}

func (cm checkCookieMiddleware) CheckCookie(
	f strictnethttp.StrictHTTPHandlerFunc,
	operationID string,
) strictnethttp.StrictHTTPHandlerFunc {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (interface{}, error) {
		if !lo.Contains([]string{"GetGamesIdRegistrations", "PostGame", "GetAuth"}, operationID) {
			return f(ctx, w, r, request)
		}

		cookie, err := r.Cookie(AuthorizationTokenName)
		if err != nil {
			return unauthorizedResponse{
				Body: api.Error{Error: err.Error()},
			}, nil
		}

		knownCookie, ok := cm.cache.Get(cookie.Value)
		if !ok {
			return unauthorizedResponse{
				Body: api.Error{Error: "unknown cookie"},
			}, nil
		}

		if knownCookie.Expires.Before(time.Now()) {
			cm.cache.Del(cookie.Value)
			return unauthorizedResponse{
				Body: api.Error{Error: "cookie is expired"},
			}, nil
		}

		//		ctx = context.WithValue(ctx, loginKey("login"), loginValue(knownCookie.Login))
		//		r = r.WithContext(ctx)

		return f(ctx, w, r, request)
	}
}

// func (cm checkCookieMiddleware) CheckCookie(next http.Handler) http.Handler {
//	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		cookie, err := r.Cookie(AuthorizationTokenName)
//		if err != nil {
//			api.MiddlewareFunc
//			resp := api.Error{Error: err.Error()}
//			w.WriteHeader(http.StatusUnauthorized)
//			err = json.NewEncoder(w).Encode(resp)
//
//			logger.Warnf("can't get cookie from request")
//			return
//		}
//
//		knownCookie, ok := cm.cache.Get(cookie.Value)
//		if !ok {
//			logger.Warnf("unknown cookie")
//			ResponseWithJSON(w, http.StatusUnauthorized, nil)
//			return
//		}
//
//		if knownCookie.Expires.Before(time.Now()) {
//			logger.Warnf("cookie is expierd")
//			cm.cache.Del(cookie.Value)
//			ResponseWithJSON(w, http.StatusUnauthorized, nil)
//			return
//		}
//
//		ctx := r.Context()
//		ctx = context.WithValue(ctx, loginKey("login"), loginValue(knownCookie.Login))
//		r = r.WithContext(ctx)
//
//		next.ServeHTTP(w, r)
//	})
// }
