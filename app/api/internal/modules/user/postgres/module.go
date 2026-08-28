package postgres

import (
	"go.uber.org/fx"

	"github.com/logstorm/api/internal/modules/user"
)

var Module = fx.Module("user/postgres",
	fx.Provide(
		fx.Annotate(
			NewPostgresUserRepository,
			fx.As(new(user.UserRepository)),
		),
	),
)
