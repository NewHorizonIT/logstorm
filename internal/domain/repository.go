package domain

import "context"

type ILogStorm interface {
	InsertLogs(ctx context.Context, logs []Log) error
}
