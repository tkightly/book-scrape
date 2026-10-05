package main

import (
	"book-scrape/book"
	"book-scrape/discord"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
)

type config struct {
	discordWebhookURL string
	searchAuthor      string
	dbPath            string
}

func loadConfig() (config, error) {
	var cfg config

	vars := []struct {
		name string
		dest *string
	}{
		{"DISCORD_WEBHOOK_URL", &cfg.discordWebhookURL},
		{"SEARCH_AUTHOR_NAME", &cfg.searchAuthor},
		{"DB_PATH", &cfg.dbPath},
	}

	var errs []error

	for _, v := range vars {
		if value := os.Getenv(v.name); value != "" {
			*v.dest = value
		} else {
			errs = append(errs, fmt.Errorf("%s is empty or not present", v.name))
		}
	}
	return cfg, errors.Join(errs...)
}

func doDiff(storedBooks, updatedBooks []book.Book) (addedBooks []book.Book, removedBooks []book.Book) {

	diffInto := func(from, against []book.Book) (out []book.Book) {

		againstMap := make(map[string]struct{})
		for _, againstBook := range against {
			againstMap[againstBook.ObjectID] = struct{}{}
		}

		// emit only distinct books
		alreadyEmittedMap := make(map[string]struct{})

		for _, fromBook := range from {
			if _, ok := againstMap[fromBook.ObjectID]; ok {
				continue
			}
			if _, ok := alreadyEmittedMap[fromBook.ObjectID]; ok {
				continue
			}
			out = append(out, fromBook)
			alreadyEmittedMap[fromBook.ObjectID] = struct{}{}
		}

		return out
	}

	removedBooks = diffInto(storedBooks, updatedBooks)

	addedBooks = diffInto(updatedBooks, storedBooks)

	return addedBooks, removedBooks
}

func main() {

	cfg, err := loadConfig()

	if err != nil {
		log.Fatalf("initialising config: %v", err)
	}

	err = discord.SendWebhook(context.Background(), []book.Book{{}}, cfg.discordWebhookURL)

	if err != nil {
		log.Fatal(err)
	}

}
