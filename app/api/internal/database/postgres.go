package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/logstorm/api/internal/config"
)

type Postgres struct {
	Pool *pgxpool.Pool
}

func Connect(conf config.DatabaseConfig) (*Postgres, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		conf.Username,
		conf.Password,
		conf.Host,
		conf.Port,
		conf.Name,
	)

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	poolConfig.MaxConns = int32(conf.MaxOpenConns)
	poolConfig.MinConns = int32(conf.MaxIdleConns)
	poolConfig.MaxConnLifetime = conf.ConnMaxLifetime
	poolConfig.MaxConnIdleTime = conf.MaxConnIdleTime
	poolConfig.HealthCheckPeriod = conf.HealthCheckPeriod

	db, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, err
	}

	return &Postgres{Pool: db}, nil
}

func (p *Postgres) Close() {
	p.Pool.Close()
}

func (p *Postgres) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return p.Pool.Ping(ctx)
}
