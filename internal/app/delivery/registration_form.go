package delivery

import (
	"net/http"

	"github.com/Snikimonkd/quizon/internal/pkg/model"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
)

func (d delivery) RegistrationForm(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_ = id
	w.WriteHeader(http.StatusOK)
	err := d.templ.ExecuteTemplate(w, "registration_form", model.RegistrationForm{})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Error().Err(err).Msg("can't execute template")
		return
	}
}
