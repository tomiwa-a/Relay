package app

import (
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tomiwa-a/Relay/internal/repository"
)

type Application struct {
	Config     Config
	Logger     *log.Logger
	Repository *repository.Queries
}

func NewApplication(config Config, logger *log.Logger, db *pgxpool.Pool) *Application {
	return &Application{
		Config:     config,
		Logger:     logger,
		Repository: repository.New(db),
	}
}
