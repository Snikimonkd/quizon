package usecase

type Repository interface {
	CheckCookieRepository
	LoginRepository
}

type usecase struct {
	checkAuthRepository CheckCookieRepository
	loginRepository     LoginRepository
}

// New - конструктор
func New(repository Repository) usecase {
	return usecase{
		checkAuthRepository: repository,
		loginRepository:     repository,
	}
}
