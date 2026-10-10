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
	"log/slog"
	"math/rand/v2"
	"os"
	"os/signal"
	"slices"
	"strings"
	"syscall"
	"time"
)

type config struct {
	discordWebhookURL string
	searchAuthors     string
	dbPath            string
	interval          string
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
		{"INTERVAL", &cfg.interval},
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

func run(ctx context.Context, cfg config) error {
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

	slog.Info("querying for books", "authors", book.BuildAuthorNames(authors))

	for _, author := range authors {
		books, err := algolia.Search(ctx, author)
		if err != nil {
			return fmt.Errorf("searching for books: %w", err)
		}

		updatedBooks = append(updatedBooks, books...)
	}

	slog.Info("books retrieved", "updatedBooks", len(updatedBooks))

	db, err := sqlite.NewDb(cfg.dbPath)
	if err != nil {
		return fmt.Errorf("initialising db: %w", err)
	}
	defer db.Close()

	storedBooks, err := db.ListBooks(ctx)
	if err != nil {
		return fmt.Errorf("getting stored books: %w", err)
	}

	slog.Info("books queried from DB", "storedBooks", len(storedBooks))

	addedBooks, removedBooks := doDiff(storedBooks, updatedBooks)

	slog.Info("books retrieved", "addedBooks", len(addedBooks), "removedBooks", len(removedBooks))

	err = db.SyncBooks(ctx, []book.Book{}, removedBooks)
	if err != nil {
		return fmt.Errorf("syncing books: %w", err)
	}

	chunkedBooks := slices.Chunk(addedBooks, discord.MaxEmbedsPerMessage)

	i := 0

	for chunk := range chunkedBooks {

		var content string

		if i == 0 {
			var plural string
			if len(addedBooks) != 1 {
				plural = "s"
			}
			content = fmt.Sprintf("Search for %s found %d book%s", book.BuildAuthorNames(authors), len(addedBooks), plural)
		}

		err = discord.SendWebhook(ctx, authors, content, chunk, cfg.discordWebhookURL)
		if err != nil {
			return fmt.Errorf("chunk %d: sending discord webhook: %w", i, err)
		}
		slog.Info("sent chunk", "chunk", i, "chunkSize", len(chunk))

		err = db.SyncBooks(ctx, chunk, []book.Book{})
		if err != nil {
			return fmt.Errorf("chunk %d: syncing books: %w", i, err)
		}
		slog.Info("synced books", "chunk", i, "chunkSize", len(chunk))

		i++
	}

	err = db.Close()
	if err != nil {
		return fmt.Errorf("closing database: %w", err)
	}

	return nil
}

func main() {
	defer slog.Info("Shutting down")

	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("initialising config: %v", err)
	}

	interval, err := time.ParseDuration(cfg.interval)
	if err != nil {
		log.Fatalf("parsing interval of %s: %v", cfg.interval, err)
	}

	if interval < 1*time.Minute {
		log.Fatalf("interval of %s is less than 1 minute", interval)
	}

	notifyCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	for {
		timeoutCtx, cancel := context.WithTimeout(notifyCtx, 1*time.Minute)

		err = run(timeoutCtx, cfg)
		cancel()

		if notifyCtx.Err() != nil {
			break
		}

		if err != nil {
			slog.Warn("A run did not complete successfully", "error", err)
		}

		wait := interval + rand.N(interval/6) - (interval / 12)

		slog.Info("next run scheduled", "in", wait)

		select {
		case <-notifyCtx.Done():
			return
		case <-time.After(wait):
		}

	}

}
