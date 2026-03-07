package clickhouse

import (
	"context"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/NewHorizonIT/logstorm/internal/config"
)

func NewClickHouse(cfg config.ClickHouseConfig) (clickhouse.Conn, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{cfg.Addr},
		Auth: clickhouse.Auth{
			Database: cfg.Database,
			Username: cfg.Username,
			Password: cfg.Password,
		},
		DialTimeout: cfg.DialTimeout,
	})

	if err != nil {
		return nil, err
	}

	ctx := context.Background()

	if err := conn.Ping(ctx); err != nil {
		return nil, err
	}

	return conn, nil
}
