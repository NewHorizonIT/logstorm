package dashboard

import "go.uber.org/fx"

var Module = fx.Module("dashboard",
	fx.Provide(
		NewDashboardService,
		NewDashboardHandler,
	),
)
