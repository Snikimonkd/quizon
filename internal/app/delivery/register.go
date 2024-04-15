package delivery

import (
	"context"
	"net/http"
	"strconv"

	"github.com/Snikimonkd/quizon/internal/pkg/model"
	"github.com/rs/zerolog/log"
)

type RegisterUsecase interface {
	Register(ctx context.Context, req model.Registration) (int64, error)
}

func (d delivery) Register(w http.ResponseWriter, r *http.Request) {
	var reg model.Registration
	reg.TeamID = r.FormValue("TeamID")
	reg.CaptainName = r.FormValue("CaptainName")
	reg.CaptainGroup = r.FormValue("CaptainGroup")
	reg.CaptainTelegram = r.FormValue("CaptainTelegram")
	reg.TeamName = r.FormValue("TeamName")

	teamSize, err := strconv.Atoi(r.FormValue("TeamSize"))
	if err != nil {
		log.Error().Err(err).Msg("can't cast team size to int")
	}
	reg.TeamSize = int64(teamSize)

	modal := model.Modal{
		Header: "header",
		Text:   "text",
		Button: "button",
	}

	w.WriteHeader(http.StatusOK)
	err = d.templ.ExecuteTemplate(w, "registration_form", model.RegistrationForm{Modal: &modal})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Error().Err(err).Msg("can't execute template")
		return
	}
}
