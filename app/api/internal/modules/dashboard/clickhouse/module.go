package clickhouse

import (
	"go.uber.org/fx"

	"github.com/logstorm/api/internal/modules/dashboard"
)

var Module = fx.Module("dashboard/clickhouse",
	fx.Provide(
		fx.Annotate(
			NewDashboardRepository,
			fx.As(new(dashboard.DashboardRepository)),
		),
	),
)
