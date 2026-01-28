package app

import (
	"flag"
	"os"
)

type Config struct {
	Port int
	Env  string
	DB   struct {
		DSN          string
		MaxOpenConns int
		MaxIdleConns int
		MaxIdleTime  string
	}
	Kafka struct {
		Brokers []string
		Topic   string
	}
}

func LoadConfig() Config {

	var config Config

	var kafkaBrokers string

	flag.IntVar(&config.Port, "port", 4000, "API server port number")
	flag.StringVar(&config.Env, "env", "development", "Environment (development|staging|production)")
	flag.StringVar(&config.DB.DSN, "db-dsn", os.Getenv("RELAY_DB_DSN"), "Database DSN")
	flag.IntVar(&config.DB.MaxOpenConns, "db-max-open-conns", 25, "Database max open connections")
	flag.IntVar(&config.DB.MaxIdleConns, "db-max-idle-conns", 25, "Database max idle connections")
	flag.StringVar(&config.DB.MaxIdleTime, "db-max-idle-time", "10m", "Database max idle time")

	flag.StringVar(&kafkaBrokers, "kafka-brokers", getEnv("RELAY_KAFKA_BROKERS", "localhost:9092"), "Kafka brokers (comma separated)")
	flag.StringVar(&config.Kafka.Topic, "kafka-topic", getEnv("RELAY_KAFKA_TOPIC", "relay-jobs"), "Kafka topic")

	flag.Parse()

	config.Kafka.Brokers = []string{kafkaBrokers}

	return config
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
