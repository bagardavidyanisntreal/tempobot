package main

import (
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	// Use an environment variable for the bot token
	botToken := os.Getenv("API_TOKEN")
	if botToken == "" {
		log.Fatal("API_TOKEN environment variable not set")
	}

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true // Enable debug output for development
	log.Printf("Authorized on account %s", bot.Self.UserName)

	// Configure updates: use long polling
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	// Loop through updates and echo messages
	for update := range updates {
		if update.Message != nil { // Ignore non-message updates
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

			// Create a new message configuration
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
			// Optionally, reply to the specific message
			msg.ReplyToMessageID = update.Message.MessageID

			// Send the echo message
			if _, err := bot.Send(msg); err != nil {
				log.Println("Error sending message:", err)
			}
		}
	}
}
