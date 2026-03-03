package model

import "time"

type LogEvent struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Service   string    `json:"service"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	TraceID   string    `json:"trace_id"`
	Env       string    `json:"env"`
}
