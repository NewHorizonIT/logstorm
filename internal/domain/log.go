package domain

type Log struct {
	ID          int64  `json:"id"`
	Message     string `json:"message" binding:"required,max=10000"`
	TraceID     string `json:"trace_id" binding:"max=200"`
	Environment string `json:"environment" binding:"required,max=50"`
	Level       string `json:"level" binding:"required,oneof=debug info warn warning error fatal DEBUG INFO WARN WARNING ERROR FATAL"`
	Service     string `json:"service" binding:"required,max=100"`
	Timestamp   int64  `json:"timestamp"`
}
