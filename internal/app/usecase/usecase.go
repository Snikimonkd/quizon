package usecase

type Repository interface {
	IndexRepository
}

type usecase struct {
	indexRepository IndexRepository
}

// NewUsecase - конструктор
func NewUsecase(repository Repository) usecase {
	return usecase{
		indexRepository: repository,
	}
}
