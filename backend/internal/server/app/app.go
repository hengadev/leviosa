package app

type App struct {
	Svcs  *Services
	Repos *Repos
}

// Function to use in the main, once all the services and repos are built.
func New(svcs *Services, repos *Repos) *App {
	return &App{
		Svcs:  svcs,
		Repos: repos,
	}
}
