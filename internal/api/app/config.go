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
}

func LoadConfig() Config {

	var config Config

	flag.IntVar(&config.Port, "port", 4000, "API server port number")
	flag.StringVar(&config.Env, "env", "development", "Environment (development|staging|production)")
	flag.StringVar(&config.DB.DSN, "db-dsn", os.Getenv("RELAY_DB_DSN"), "Database DSN")
	flag.IntVar(&config.DB.MaxOpenConns, "db-max-open-conns", 25, "Database max open connections")
	flag.IntVar(&config.DB.MaxIdleConns, "db-max-idle-conns", 25, "Database max idle connections")
	flag.StringVar(&config.DB.MaxIdleTime, "db-max-idle-time", "10m", "Database max idle time")

	flag.Parse()

	return config
}
