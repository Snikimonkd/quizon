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
	gameID, err := strconv.Atoi(r.FormValue("GameID"))
	if err != nil {
		log.Error().Err(err).Msg("can't cast game id to int")
	}
	reg.GameID = int64(gameID)

	_, err = d.registerUsecase.Register(r.Context(), reg)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Error().Err(err).Msg("")
		return
	}

	modal := model.Modal{
		Header: "заголовок",
		Text:   "вы там чето справились с первым заданием хуе мое так держать",
		Button: "кнопка",
	}

	w.WriteHeader(http.StatusOK)
	err = d.templ.ExecuteTemplate(w, "modal", modal)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Error().Err(err).Msg("can't execute template")
		return
	}
}
