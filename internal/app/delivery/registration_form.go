package delivery

import (
	"net/http"

	"github.com/rs/zerolog/log"
)

func (d delivery) RegistrationForm(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	err := d.templ.ExecuteTemplate(w, "registration_form", struct{}{})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Error().Err(err).Msg("can't execute template")
		return
	}
}
