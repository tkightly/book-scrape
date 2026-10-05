package main

import (
	"book-scrape/algolia"
	"book-scrape/book"
	"book-scrape/discord"
	"book-scrape/sqlite"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
)

type config struct {
	discordWebhookURL string
	searchAuthors     string
	dbPath            string
}

func loadConfig() (config, error) {
	var cfg config

	vars := []struct {
		name string
		dest *string
	}{
		{"DISCORD_WEBHOOK_URL", &cfg.discordWebhookURL},
		{"SEARCH_AUTHOR_NAMES", &cfg.searchAuthors},
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

	authorsSlice := strings.Split(cfg.searchAuthors, ",")
	var authors []string
	for i := range authorsSlice {
		trimmed := strings.TrimSpace(authorsSlice[i])
		if trimmed != "" {
			authors = append(authors, authorsSlice[i])
		}
	}

	if len(authors) < 1 {
		log.Fatalf("no authors to be searched in SEARCH_AUTHOR_NAMES: %q", cfg.searchAuthors)
	}

	ctx := context.Background()

	var updatedBooks []book.Book

	for _, author := range authors {
		books, err := algolia.Search(ctx, author)
		if err != nil {
			log.Fatalf("searching for books: %v", err)
		}

		updatedBooks = append(updatedBooks, books...)
	}

	db, err := sqlite.NewDb(cfg.dbPath)
	if err != nil {
		log.Fatalf("initialising db: %v", err)
	}

	storedBooks, err := db.ListBooks(ctx)
	if err != nil {
		log.Fatalf("getting stored books: %v", err)
	}

	addedBooks, removedBooks := doDiff(storedBooks, updatedBooks)

	err = discord.SendWebhook(ctx, authors, addedBooks, cfg.discordWebhookURL)
	if err != nil {
		log.Fatalf("sending discord webhook: %v", err)
	}

	err = db.SyncBooks(ctx, addedBooks, removedBooks)
	if err != nil {
		log.Fatalf("syncing books: %v", err)
	}

	err = db.Close()
	if err != nil {
		log.Fatalf("closing database: %v", err)
	}

}
