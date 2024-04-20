package usecase

type Repository interface {
	CheckCookieRepository
	LoginRepository
	ListGamesRepository
	RegisterRepository
}

type usecase struct {
	checkAuthRepository CheckCookieRepository
	loginRepository     LoginRepository
	listGamesRepository ListGamesRepository
	registerRepository  RegisterRepository
}

// New - конструктор
func New(repository Repository) usecase {
	return usecase{
		checkAuthRepository: repository,
		loginRepository:     repository,
		listGamesRepository: repository,
		registerRepository:  repository,
	}
}
