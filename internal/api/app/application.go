package app

import (
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Application struct {
	Config     Config
	Logger     *log.Logger
	Connection *pgxpool.Pool
}

func NewApplication(config Config, logger *log.Logger, connection *pgxpool.Pool) *Application {
	return &Application{
		Config:     config,
		Logger:     logger,
		Connection: connection,
	}
}
