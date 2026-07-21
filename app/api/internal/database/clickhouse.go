package database

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/logstorm/api/internal/config"
)

type ClickHouseClient interface {
	Query(ctx context.Context, query string, args ...interface{}) (driver.Rows, error)
	QueryRow(ctx context.Context, query string, args ...interface{}) driver.Row
	Exec(ctx context.Context, query string, args ...interface{}) error
	PrepareBatch(ctx context.Context, query string) (driver.Batch, error)
	Ping() error
	HealthCheck() error
	Close() error
}

type ClickHouse struct {
	Conn clickhouse.Conn
}

func (ch *ClickHouse) Exec(ctx context.Context, query string, args ...interface{}) error {
	return ch.Conn.Exec(ctx, query, args...)
}

func (ch *ClickHouse) PrepareBatch(ctx context.Context, query string) (driver.Batch, error) {
	return ch.Conn.PrepareBatch(ctx, query)
}

func (ch *ClickHouse) QueryRow(ctx context.Context, query string, args ...interface{}) driver.Row {
	return ch.Conn.QueryRow(ctx, query, args...)
}

func (ch *ClickHouse) Query(ctx context.Context, query string, args ...interface{}) (driver.Rows, error) {
	return ch.Conn.Query(ctx, query, args...)
}

func NewClickHouse(conf config.ClickHouseConfig) (ClickHouseClient, error) {
	var protocol clickhouse.Protocol
	switch conf.Protocol {
	case "native":
		protocol = clickhouse.Native
	case "http":
		protocol = clickhouse.HTTP
	default:
		return nil, fmt.Errorf("invalid ClickHouse protocol: %s", conf.Protocol)
	}

	conn, err := clickhouse.Open(&clickhouse.Options{
		Protocol: protocol,
		Addr:     []string{conf.Host + ":" + strconv.Itoa(conf.HTTPPort)},
		Auth: clickhouse.Auth{
			Database: conf.Database,
			Username: conf.Username,
			Password: conf.Password,
		},
		MaxOpenConns:    conf.MaxOpenConns,
		MaxIdleConns:    conf.MaxIdleConns,
		ConnMaxLifetime: conf.ConnMaxLifetime,
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
	})
	if err != nil {
		return nil, err
	}

	return &ClickHouse{
		Conn: conn,
	}, nil
}

func (ch *ClickHouse) Close() error {
	return ch.Conn.Close()
}

func (ch *ClickHouse) Ping() error {
	ctx, cancle := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancle()
	return ch.Conn.Ping(ctx)
}

func (ch *ClickHouse) HealthCheck() error {
	if err := ch.Ping(); err != nil {
		return fmt.Errorf("clickhouse health check failed: %w", err)
	}

	// Select 1
	var result int
	ctx, cancle := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancle()

	if err := ch.Conn.QueryRow(ctx, "SELECT 1").Scan(&result); err != nil {
		return fmt.Errorf("clickhouse health check failed: %w", err)
	}
	if result != 1 {
		return fmt.Errorf("clickhouse health check failed: expected 1, got %d", result)
	}
	return nil
}
