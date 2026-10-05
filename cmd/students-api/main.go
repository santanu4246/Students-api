package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
    "github.com/santanu/students-api/internal"
	"github.com/santanu/students-api/internal/http/handlers/student"
)

func main() {
	//load config
	cfg := config.MustLoad()
	//database setup
	//setup router
	router := http.NewServeMux()

	router.HandleFunc("POST /api/students", student.New())
	//setup server
	server := http.Server{
		Addr:    cfg.Addr,
		Handler: router,
	}

	slog.Info("server started", slog.String("address:", cfg.Addr))
	fmt.Printf("Server Started %s", cfg.Addr)

	done := make(chan os.Signal, 1)

	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	
	go func() {
		err := server.ListenAndServe()
		if err != nil {
			log.Fatal("Failed to start the server")
		}
	}()
	
	<-done
	
	slog.Info("shuting down the server")

	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()

	err:= server.Shutdown(ctx)

	if err != nil {
		slog.Error("failed to shutdown server", slog.String("error", err.Error()))
	}

	slog.Info("Server shutdown sucessfully")

}
