package main

import (
	"github.com/logstorm/api/internal/bootstrap"
)

func main() {
	app, err := bootstrap.NewApp()
	if err != nil {
		panic(err)
	}

	defer app.Logger.Close()

}
