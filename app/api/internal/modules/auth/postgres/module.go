package postgres

import (
	"go.uber.org/fx"

	"github.com/logstorm/api/internal/modules/auth"
)

var Module = fx.Module("auth/postgres",
	fx.Provide(
		fx.Annotate(
			NewPostgresAuthRepository,
			fx.As(new(auth.AuthRepository)),
		),
	),
)
