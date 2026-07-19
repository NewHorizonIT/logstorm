package main

import (
	"log"
	"net"
	"strconv"

	"github.com/logstorm/api/internal/bootstrap"
)

func main() {
	app, err := bootstrap.NewApp()
	defer app.Logger.Close()

	if err != nil {
		log.Fatal(err)
	}

	address := net.JoinHostPort(app.Config.Server.Host, strconv.Itoa(app.Config.Server.Port))
	// address := fmt.Sprintf("%s:%d", app.Config.Server.Host, app.Config.Server.Port)

	if err := app.Router.Run(address); err != nil {
		app.Logger.Zerolog.Error().Err(err).Msg("Failed to start server")
	}

}
