package app

import (
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/segmentio/kafka-go"
	"github.com/tomiwa-a/Relay/internal/repository"
)

type Application struct {
	Config      Config
	Logger      *log.Logger
	Repository  *repository.Queries
	KafkaWriter *kafka.Writer
}

func NewApplication(config Config, logger *log.Logger, db *pgxpool.Pool, kafkaWriter *kafka.Writer) *Application {
	return &Application{
		Config:      config,
		Logger:      logger,
		Repository:  repository.New(db),
		KafkaWriter: kafkaWriter,
	}
}
