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
	"time"
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

func run(ctx context.Context) error {
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("initialising config: %w", err)
	}

	authorsSlice := strings.Split(cfg.searchAuthors, ",")
	var authors []string
	for i := range authorsSlice {
		trimmed := strings.TrimSpace(authorsSlice[i])
		if trimmed != "" {
			authors = append(authors, trimmed)
		}
	}

	if len(authors) < 1 {
		return fmt.Errorf("no authors to be searched in SEARCH_AUTHOR_NAMES: %q", cfg.searchAuthors)
	}

	var updatedBooks []book.Book

	for _, author := range authors {
		books, err := algolia.Search(ctx, author)
		if err != nil {
			return fmt.Errorf("searching for books: %w", err)
		}

		updatedBooks = append(updatedBooks, books...)
	}

	db, err := sqlite.NewDb(cfg.dbPath)
	if err != nil {
		return fmt.Errorf("initialising db: %w", err)
	}
	defer db.Close()

	storedBooks, err := db.ListBooks(ctx)
	if err != nil {
		return fmt.Errorf("getting stored books: %w", err)
	}

	addedBooks, removedBooks := doDiff(storedBooks, updatedBooks)

	err = discord.SendWebhook(ctx, authors, addedBooks, cfg.discordWebhookURL)
	if err != nil {
		return fmt.Errorf("sending discord webhook: %w", err)
	}

	err = db.SyncBooks(ctx, addedBooks, removedBooks)
	if err != nil {
		return fmt.Errorf("syncing books: %w", err)
	}

	err = db.Close()
	if err != nil {
		return fmt.Errorf("closing database: %w", err)
	}

	return nil
}

func main() {

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	err := run(ctx)
	if err != nil {
		log.Fatal(err)
	}

}
