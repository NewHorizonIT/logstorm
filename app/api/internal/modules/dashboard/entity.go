package dashboard

import "time"

type VolumePoint struct {
	Time  time.Time `json:"time"`
	Count int64     `json:"count"`
}

type LevelBreakdown struct {
	Level string `json:"level"`
	Count int64  `json:"count"`
}

type RecentError struct {
	Timestamp   time.Time `json:"timestamp"`
	Level       string    `json:"level"`
	Service     string    `json:"service"`
	Message     string    `json:"message"`
	Environment string    `json:"environment"`
}

type LogStats struct {
	TotalLogs      int64            `json:"total_logs"`
	ErrorCount     int64            `json:"error_count"`
	WarnCount      int64            `json:"warn_count"`
	Volume         []VolumePoint    `json:"volume"`
	LevelBreakdown []LevelBreakdown `json:"level_breakdown"`
	RecentErrors   []RecentError    `json:"recent_errors"`
	From           time.Time        `json:"from"`
	To             time.Time        `json:"to"`
}
