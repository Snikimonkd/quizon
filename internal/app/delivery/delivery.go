package delivery

import (
	"fmt"
	"net/http"
	"text/template"

	"github.com/rs/zerolog/log"
)

type delivery struct {
	indexUsecase IndexUsecase

	templ *template.Template
}

// New - конструктор
func New(
	indexUsecase IndexUsecase,
) (delivery, error) {
	templ, err := template.ParseGlob("templates/*.html")
	if err != nil {
		return delivery{}, fmt.Errorf("can't parse templates: %w", err)
	}

	return delivery{
		indexUsecase: indexUsecase,
		templ:        templ,
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
