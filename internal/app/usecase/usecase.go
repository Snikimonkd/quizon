package usecase

type Repository interface {
	CheckCookieRepository
	LoginRepository
	ListGamesRepository
}

type usecase struct {
	checkAuthRepository CheckCookieRepository
	loginRepository     LoginRepository
	listGamesRepository ListGamesRepository
}

// New - конструктор
func New(repository Repository) usecase {
	return usecase{
		checkAuthRepository: repository,
		loginRepository:     repository,
		listGamesRepository: repository,
	}
}
