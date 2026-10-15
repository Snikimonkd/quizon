package delivery

type Usecase interface {
	ListGamesUsecase
	ListRegistrationsUsecase

	CreateRegistrationUsecase
	CreateGameUsecase

	LoginUsecase
}

type delivery struct {
	usecase Usecase
}

func New(usecase Usecase) delivery {
	return delivery{
		usecase: usecase,
	}
}
