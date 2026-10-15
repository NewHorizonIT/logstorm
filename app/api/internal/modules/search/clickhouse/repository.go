package clickhouse

import (
	"context"
	"fmt"
	"strings"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/logstorm/api/internal/modules/search"
)

type LogRepository struct {
	conn driver.Conn
}

func NewLogRepository(conn driver.Conn) *LogRepository {
	return &LogRepository{conn: conn}
}

func (r *LogRepository) Search(ctx context.Context, q search.SearchQuery) (*search.SearchResult, error) {
	where, args := buildWhereClause(q)

	// Count total
	countSQL := fmt.Sprintf("SELECT count() FROM logs %s", where)
	row := r.conn.QueryRow(ctx, countSQL, args...)
	var total int64
	if err := row.Scan(&total); err != nil {
		return nil, fmt.Errorf("count logs: %w", err)
	}

	// Fetch logs
	querySQL := fmt.Sprintf(`
        SELECT id, timestamp, project_id, environment, level, service, message,
               trace_id, span_id, source, attributes
        FROM logs
        %s
        ORDER BY timestamp DESC
        LIMIT ? OFFSET ?`,
		where,
	)
	args = append(args, q.Limit, q.Offset)

	rows, err := r.conn.Query(ctx, querySQL, args...)
	if err != nil {
		return nil, fmt.Errorf("query logs: %w", err)
	}
	defer rows.Close()

	var logs []*search.LogRecord
	for rows.Next() {
		var log search.LogRecord
		if err := rows.ScanStruct(&log); err != nil {
			return nil, fmt.Errorf("scan log row: %w", err)
		}
		logs = append(logs, &log)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate log rows: %w", err)
	}

	return &search.SearchResult{Logs: logs, Total: total}, nil
}

func buildWhereClause(q search.SearchQuery) (string, []any) {
	conditions := []string{
		"project_id = ?",
		"timestamp >= ?",
		"timestamp < ?",
	}
	args := []any{q.ProjectID, q.From, q.To}

	if q.Level != "" {
		conditions = append(conditions, "level = ?")
		args = append(args, q.Level)
	}
	if q.Service != "" {
		conditions = append(conditions, "service = ?")
		args = append(args, q.Service)
	}
	if q.Text != "" {
		conditions = append(conditions, "message ILIKE ?")
		args = append(args, "%"+q.Text+"%")
	}

	return "WHERE " + strings.Join(conditions, " AND "), args
}
