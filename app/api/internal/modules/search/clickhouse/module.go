package clickhouse

import (
	"go.uber.org/fx"

	"github.com/logstorm/api/internal/modules/search"
)

var Module = fx.Module("search/clickhouse",
	fx.Provide(
		fx.Annotate(
			NewLogRepository,
			fx.As(new(search.LogRepository)),
		),
	),
)
