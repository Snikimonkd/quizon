package delivery

import (
	"fmt"
	"net/http"
	"text/template"

	"github.com/rs/zerolog/log"
)

type Usecase interface {
	CheckCookieUsecase
	LoginUsecase
	ListGamesUsecase
	RegisterUsecase
}

type delivery struct {
	checkAuthUsecase CheckCookieUsecase
	loginUsecase     LoginUsecase
	listGamesUsecase ListGamesUsecase
	registerUsecase  RegisterUsecase

	templ *template.Template
}

// New - конструктор
func New(
	usecase Usecase,
) (delivery, error) {
	templ, err := template.ParseGlob("./front/templates/*.html")
	if err != nil {
		return delivery{}, fmt.Errorf("can't parse templates: %w", err)
	}

	return delivery{
		checkAuthUsecase: usecase,
		loginUsecase:     usecase,
		listGamesUsecase: usecase,
		registerUsecase:  usecase,

		templ: templ,
	}, nil
}

func writeAll(w http.ResponseWriter, msg []byte) error {
	n := 0
	for n != len(msg) {
		m, err := w.Write(msg[n:])
		if err != nil {
			return fmt.Errorf("can't write all: %w", err)
		}
		n += m
	}

	return nil
}

func ResponseWithError(code int, text string, w http.ResponseWriter) {
	w.WriteHeader(code)
	err := writeAll(w, []byte(text))
	if err != nil {
		w.WriteHeader(500)
		log.Error().Err(err).Msg("")
	}
}
