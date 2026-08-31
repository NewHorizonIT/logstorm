package main

import (
	"go.uber.org/fx"

	"github.com/logstorm/api/internal/bootstrap"
	"github.com/logstorm/api/internal/modules/auth"
	authpostgres "github.com/logstorm/api/internal/modules/auth/postgres"
	"github.com/logstorm/api/internal/modules/user"
	userpostgres "github.com/logstorm/api/internal/modules/user/postgres"
)

func main() {
	fx.New(
		bootstrap.Module,
		user.Module,
		userpostgres.Module,
		auth.Module,
		authpostgres.Module,
	).Run()
}
