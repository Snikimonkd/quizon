package usecase

type Repository interface {
	IndexRepository
}

type usecase struct {
	indexRepository IndexRepository
}

// New - конструктор
func New(repository Repository) usecase {
	return usecase{
		indexRepository: repository,
	}
}
