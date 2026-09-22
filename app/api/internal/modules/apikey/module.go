package apikey

import "go.uber.org/fx"

var Module = fx.Module("apikey",
	fx.Provide(
		NewAPIKeyService,
		NewAPIKeyHandler,
		NewMiddleware,
	),
)
