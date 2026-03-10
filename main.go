package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bagardavidyanisntreal/tempobot/internal/service"
	"github.com/bagardavidyanisntreal/tempobot/internal/storage"
	"github.com/bagardavidyanisntreal/tempobot/internal/telegram"

	_ "github.com/lib/pq"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	token := os.Getenv("API_TOKEN")
	if len(token) == 0 {
		log.Printf("API_TOKEN env variable not set")

		return
	}

	dbURL := os.Getenv("DB_URL")
	if len(dbURL) == 0 {
		log.Printf("DB_URL env variable not set")

		return
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Printf("failed get DB connect: %s", err)

		return
	}

	if err = db.PingContext(ctx); err != nil {
		log.Printf("failed ping DB: %s", err)

		return
	}

	userRepo := storage.NewUserRepo(db)
	eventRepo := storage.NewEventRepo(db)
	partRepo := storage.NewParticipantRepo(db)

	userService := service.NewUserService(userRepo)
	eventService := service.NewEventService(eventRepo, partRepo)

	tg := telegram.NewClient(token)

	tgHandler := telegram.NewHandler(userService, eventService, tg)

	mux := http.NewServeMux()

	mux.HandleFunc("/live", func(w http.ResponseWriter, r *http.Request) {
		if err = db.PingContext(r.Context()); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}

		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/webhook", tgHandler.HandleWebhook)

	log.Println("server started :8080")

	//nolint:exhaustruct,mnd
	server := &http.Server{
		Addr:        ":8080",
		Handler:     mux,
		ReadTimeout: 5 * time.Second,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server listen: %s", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down...")

	//nolint: mnd
	sdCtx, sdCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer sdCancel()

	if err := server.Shutdown(sdCtx); err != nil {
		log.Printf("server shutdown failed: %s", err)

		return
	}

	if err := db.Close(); err != nil {
		log.Printf("db close error: %s", err)
	}
}
