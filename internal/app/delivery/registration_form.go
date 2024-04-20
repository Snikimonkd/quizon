package delivery

import (
	"net/http"
	"strconv"

	"github.com/Snikimonkd/quizon/internal/pkg/model"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
)

func (d delivery) RegistrationForm(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Error().Err(err).Msg("can't get game id from url")
		return
	}

	w.WriteHeader(http.StatusOK)
	err = d.templ.ExecuteTemplate(w, "registration_form", model.RegistrationForm{GameID: id})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Error().Err(err).Msg("can't execute template")
		return
	}
}
