package app

import (
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"github.com/tomiwa-a/Relay/internal/repository"
)

type Application struct {
	Config      Config
	Logger      *log.Logger
	Repository  *repository.Queries
	KafkaWriter *kafka.Writer
	Redis       *redis.Client
}

func NewApplication(config Config, logger *log.Logger, db *pgxpool.Pool, kafkaWriter *kafka.Writer, redisClient *redis.Client) *Application {
	return &Application{
		Config:      config,
		Logger:      logger,
		Repository:  repository.New(db),
		KafkaWriter: kafkaWriter,
		Redis:       redisClient,
	}
}
