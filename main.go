package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/bagardavidyanisntreal/tempobot/internal/service"
	"github.com/bagardavidyanisntreal/tempobot/internal/storage"
	"github.com/bagardavidyanisntreal/tempobot/internal/telegram"

	_ "github.com/lib/pq"
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

	//nolint:exhaustruct,mnd
	server := &http.Server{
		Addr:              ":8080",
		ReadTimeout:       5 * time.Second,  // время на чтение заголовков и тела
		WriteTimeout:      10 * time.Second, // время на запись ответа
		IdleTimeout:       15 * time.Second, // время между запросами (keep-alive)
		ReadHeaderTimeout: 2 * time.Second,  // ограничение времени на чтение заголовков
	}

	log.Println("server started :8080")

	// Используем метод сервера вместо глобальной функции
	err = server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
