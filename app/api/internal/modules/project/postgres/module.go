package postgres

import (
	"go.uber.org/fx"

	"github.com/logstorm/api/internal/modules/project"
)

var Module = fx.Module("project/postgres",
	fx.Provide(
		fx.Annotate(
			NewPostgresProjectRepository,
			fx.As(new(project.ProjectRepository)),
		),
	),
)
