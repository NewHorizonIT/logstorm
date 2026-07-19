package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/logstorm/api/internal/bootstrap"
)

func main() {
	app, err := bootstrap.NewApp()
	if err != nil {
		log.Fatal(err)
	}

	address := net.JoinHostPort(app.Config.Server.Host, strconv.Itoa(app.Config.Server.Port))
	// address := fmt.Sprintf("%s:%d", app.Config.Server.Host, app.Config.Server.Port)

	srv := &http.Server{
		Addr:    address,
		Handler: app.Router,
	}

	// Goroutine to start the server
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Channel to receive OS signals
	sig := make(chan os.Signal, 1)

	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	// Main goroutine waits for a signal
	<-sig

	log.Println("Shutting down server...")
	// Context timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server shutdown: %s\n", err)
	}

	// App Close
	if err := bootstrap.CloseApp(app); err != nil {
		log.Fatalf("failed to close app: %s\n", err)
	}
}
