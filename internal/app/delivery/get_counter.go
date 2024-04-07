package delivery

import (
	"context"
	"net/http"

	"github.com/Snikimonkd/quizon/internal/pkg/model"
	"github.com/rs/zerolog/log"
)

type GetCoutnerUsecase interface {
	GetCounter(ctx context.Context) (int64, error)
}

func (d delivery) GetCounter(w http.ResponseWriter, r *http.Request) {
	c, err := d.getCoutnerUsecase.GetCounter(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Error().Err(err).Msg("")
		return
	}

	w.WriteHeader(http.StatusOK)
	err = d.templ.ExecuteTemplate(w, "index", model.IndexResponse{Count: c})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Error().Err(err).Msg("can't execute template")
		return
	}
}
