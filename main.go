package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/bagardavidyanisntreal/tempobot/internal/service"
	"github.com/bagardavidyanisntreal/tempobot/internal/storage"
	"github.com/bagardavidyanisntreal/tempobot/internal/telegram"
)

func main() {

	token := os.Getenv("BOT_TOKEN")
	dbURL := os.Getenv("DB_URL")

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}

	userRepo := storage.NewUserRepo(db)
	eventRepo := storage.NewEventRepo(db)
	partRepo := storage.NewParticipantRepo(db)

	userService := service.NewUserService(userRepo)
	eventService := service.NewEventService(eventRepo, partRepo)

	tg := telegram.NewClient(token)

	handler := telegram.NewHandler(userService, eventService, tg)

	http.HandleFunc("/webhook", handler.HandleWebhook)

	log.Println("server started :8080")

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
