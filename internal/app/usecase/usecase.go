package usecase

type Repository interface {
	GetCounterRepository
	SetCounterRepository
}

type usecase struct {
	getCounterRepository GetCounterRepository
	setCounterRepository SetCounterRepository
}

// New - конструктор
func New(repository Repository) usecase {
	return usecase{
		getCounterRepository: repository,
		setCounterRepository: repository,
	}
}
