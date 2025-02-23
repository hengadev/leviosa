package app

type App struct {
	Svcs  *Services
	Repos *Repos
}

func New(svcs *Services, repos *Repos) *App {
	return &App{
		Svcs:  svcs,
		Repos: repos,
	}
}
