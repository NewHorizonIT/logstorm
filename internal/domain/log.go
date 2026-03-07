package domain

type Log struct {
	ID          int64  `json:"id"`
	Message     string `json:"message"`
	TraceID     string `json:"trace_id"`
	Environment string `json:"environment"`
	Level       string `json:"level"`
	Service     string `json:"service"`
	Timestamp   int64  `json:"timestamp"`
}
