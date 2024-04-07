package usecase

import "context"

type SetCounterRepository interface {
	SetCounter(ctx context.Context) (int64, error)
}

func (u usecase) SetCounter(ctx context.Context) (int64, error) {
	return u.setCounterRepository.SetCounter(ctx)
}
