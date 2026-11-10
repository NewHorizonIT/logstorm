package search

import "go.uber.org/fx"

var Module = fx.Module("search",
	fx.Provide(
		NewSearchService,
		NewSearchHandler,
	),
)
