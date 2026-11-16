package clickhouse

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/logstorm/api/internal/database"
	"github.com/logstorm/api/internal/modules/dashboard"
)

const recentErrorsLimit = 10

const queryVolumeHourly = `
	SELECT toStartOfHour(timestamp) AS bucket, count() AS cnt
	FROM logs
	WHERE project_id = ? AND timestamp >= ? AND timestamp < ?
	GROUP BY bucket
	ORDER BY bucket ASC`

const queryVolumeDaily = `
	SELECT toStartOfDay(timestamp) AS bucket, count() AS cnt
	FROM logs
	WHERE project_id = ? AND timestamp >= ? AND timestamp < ?
	GROUP BY bucket
	ORDER BY bucket ASC`

type ClickHouseDashboardRepository struct {
	conn database.ClickHouseClient
}

func NewDashboardRepository(conn database.ClickHouseClient) *ClickHouseDashboardRepository {
	return &ClickHouseDashboardRepository{conn: conn}
}

func (r *ClickHouseDashboardRepository) GetStats(ctx context.Context, projectID uuid.UUID, from, to time.Time) (*dashboard.LogStats, error) {
	stats := &dashboard.LogStats{From: from, To: to}

	// 1. Total + error + warn counts
	row := r.conn.QueryRow(ctx, `
		SELECT
			count() AS total,
			countIf(level = 'error') AS errors,
			countIf(level = 'warn') AS warns
		FROM logs
		WHERE project_id = ? AND timestamp >= ? AND timestamp < ?`,
		projectID, from, to,
	)
	if err := row.Scan(&stats.TotalLogs, &stats.ErrorCount, &stats.WarnCount); err != nil {
		return nil, fmt.Errorf("get stats counts: %w", err)
	}

	// 2. Volume — hourly buckets for <=7 days, daily for longer ranges
	volumeSQL := queryVolumeHourly
	if to.Sub(from) > 7*24*time.Hour {
		volumeSQL = queryVolumeDaily
	}

	volumeRows, err := r.conn.Query(ctx, volumeSQL, projectID, from, to)
	if err != nil {
		return nil, fmt.Errorf("get volume: %w", err)
	}
	defer volumeRows.Close()

	stats.Volume = make([]dashboard.VolumePoint, 0)
	for volumeRows.Next() {
		var p dashboard.VolumePoint
		if err := volumeRows.Scan(&p.Time, &p.Count); err != nil {
			return nil, fmt.Errorf("scan volume row: %w", err)
		}
		stats.Volume = append(stats.Volume, p)
	}
	if err := volumeRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate volume rows: %w", err)
	}

	// 3. Level breakdown
	levelRows, err := r.conn.Query(ctx, `
		SELECT level, count() AS cnt
		FROM logs
		WHERE project_id = ? AND timestamp >= ? AND timestamp < ?
		GROUP BY level
		ORDER BY cnt DESC`,
		projectID, from, to,
	)
	if err != nil {
		return nil, fmt.Errorf("get level breakdown: %w", err)
	}
	defer levelRows.Close()

	stats.LevelBreakdown = make([]dashboard.LevelBreakdown, 0)
	for levelRows.Next() {
		var lb dashboard.LevelBreakdown
		if err := levelRows.Scan(&lb.Level, &lb.Count); err != nil {
			return nil, fmt.Errorf("scan level row: %w", err)
		}
		stats.LevelBreakdown = append(stats.LevelBreakdown, lb)
	}
	if err := levelRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate level rows: %w", err)
	}

	// 4. Recent errors — last 10 error/warn logs
	recentRows, err := r.conn.Query(ctx, `
		SELECT timestamp, level, service, message, environment
		FROM logs
		WHERE project_id = ? AND timestamp >= ? AND timestamp < ?
		  AND level IN ('error', 'warn')
		ORDER BY timestamp DESC
		LIMIT ?`,
		projectID, from, to, recentErrorsLimit,
	)
	if err != nil {
		return nil, fmt.Errorf("get recent errors: %w", err)
	}
	defer recentRows.Close()

	stats.RecentErrors = make([]dashboard.RecentError, 0)
	for recentRows.Next() {
		var re dashboard.RecentError
		if err := recentRows.Scan(&re.Timestamp, &re.Level, &re.Service, &re.Message, &re.Environment); err != nil {
			return nil, fmt.Errorf("scan recent error row: %w", err)
		}
		stats.RecentErrors = append(stats.RecentErrors, re)
	}
	if err := recentRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate recent error rows: %w", err)
	}

	return stats, nil
}
