package postgres

import (
	"go.uber.org/fx"

	"github.com/logstorm/api/internal/modules/apikey"
)

var Module = fx.Module("apikey/postgres",
	fx.Provide(
		fx.Annotate(
			NewPostgresAPIKeyRepository,
			fx.As(new(apikey.APIKeyRepository)),
		),
	),
)
