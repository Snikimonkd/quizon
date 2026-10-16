package web

// page - страница сайта; значение совпадает с именем шаблона в templates/pages без .html
type page string

const (
	pageHome     page = "home"
	pageRegister page = "register"
	pageLogin    page = "login"
	pageGameNew  page = "game_new"
	pageTeams    page = "teams"
)

// allPages - все страницы. parsePages проверяет, что у каждой есть шаблон и что лишних шаблонов нет
var allPages = []page{pageHome, pageRegister, pageLogin, pageGameNew, pageTeams}

func (p page) templatePath() string {
	return "templates/pages/" + string(p) + ".html"
}

// pageTitle - заголовок страницы во вкладке браузера
type pageTitle string

const (
	titleHome     pageTitle = "Игры"
	titleRegister pageTitle = "Регистрация на игру"
	titleLogin    pageTitle = "Вход"
	titleGameNew  pageTitle = "Новая игра"
	titleTeams    pageTitle = "Участники"
)

// title - у каждой страницы заголовок один, поэтому он не передаётся при рендере отдельно
func (p page) title() pageTitle {
	switch p {
	case pageHome:
		return titleHome
	case pageRegister:
		return titleRegister
	case pageLogin:
		return titleLogin
	case pageGameNew:
		return titleGameNew
	case pageTeams:
		return titleTeams
	}

	return ""
}
