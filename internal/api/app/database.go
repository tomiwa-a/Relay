package app

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DBConfig struct {
	DSN          string
	MaxOpenConns int
	MaxIdleConns int
	MaxIdleTime  string
}

func OpenDB(cfg DBConfig) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(cfg.DSN)
	if err != nil {
		return nil, err
	}

	config.MaxConns = int32(cfg.MaxOpenConns)
	config.MinConns = int32(cfg.MaxIdleConns)
	config.MaxConnIdleTime, _ = time.ParseDuration(cfg.MaxIdleTime)

	return pgxpool.NewWithConfig(context.Background(), config)
}
