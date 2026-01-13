package app

import "log"

type Application struct {
	Config Config
	Logger *log.Logger
}

func NewApplication(config Config, logger *log.Logger) *Application {
	return &Application{
		Config: config,
		Logger: logger,
	}
}
