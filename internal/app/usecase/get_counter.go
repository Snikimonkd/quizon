package usecase

import "context"

type GetCounterRepository interface {
	GetCounter(ctx context.Context) (int64, error)
}

func (u usecase) GetCounter(ctx context.Context) (int64, error) {
	return u.getCounterRepository.GetCounter(ctx)
}
