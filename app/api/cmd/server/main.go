package main

import (
	"go.uber.org/fx"

	"github.com/logstorm/api/internal/bootstrap"
	"github.com/logstorm/api/internal/modules/apikey"
	apikeypostgres "github.com/logstorm/api/internal/modules/apikey/postgres"
	"github.com/logstorm/api/internal/modules/auth"
	authpostgres "github.com/logstorm/api/internal/modules/auth/postgres"
	"github.com/logstorm/api/internal/modules/project"
	projectpostgres "github.com/logstorm/api/internal/modules/project/postgres"
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
		project.Module,
		projectpostgres.Module,
		apikey.Module,
		apikeypostgres.Module,
	).Run()
}
