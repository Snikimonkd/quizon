package usecase

import "context"

type IndexRepository interface {
	Inc(ctx context.Context) (int64, error)
}

func (u usecase) Inc(ctx context.Context) (int64, error) {
	return u.indexRepository.Inc(ctx)
}
