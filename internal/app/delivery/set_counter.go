package delivery

import (
	"context"
	"net/http"

	"github.com/Snikimonkd/quizon/internal/pkg/model"
	"github.com/rs/zerolog/log"
)

type SetCoutnerUsecase interface {
	SetCounter(ctx context.Context) (int64, error)
}

func (d delivery) SetCounter(w http.ResponseWriter, r *http.Request) {
	c, err := d.setCoutnerUsecase.SetCounter(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Error().Err(err).Msg("")
		return
	}

	w.WriteHeader(http.StatusOK)
	err = d.templ.ExecuteTemplate(w, "count", model.IndexResponse{Count: c})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Error().Err(err).Msg("can't execute template")
		return
	}
}
