package usecase

import (
	"context"
	"time"

	"github.com/Snikimonkd/quizon/internal/pkg/model"
)

type RegisterRepository interface {
	Register(ctx context.Context, registration model.Registration) error
}

func (u usecase) Register(ctx context.Context, req model.Registration) (int64, error) {
	req.CreatedAt = time.Now()

	err := u.registerRepository.Register(ctx, req)
	if err != nil {
		return 0, err
	}

	return 0, nil
}
