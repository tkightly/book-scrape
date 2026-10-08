package discord

import (
	"book-scrape/book"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"
)

type webhookBody struct {
	Content string  `json:"content,omitempty"`
	Embeds  []embed `json:"embeds"`
}

type embed struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Color       int    `json:"color"`
	Image       *image `json:"image,omitempty"`
}

type image struct {
	URL string `json:"url"`
}

const responseBodyReadLimit int64 = 4096

const MaxEmbedsPerMessage int64 = 10

func buildAuthorNames(authorList []string) (authorString string) {

	switch len(authorList) {
	case 0:
		return ""
	case 1:
		return authorList[0]
	default:
		authorString = strings.Join(authorList[:len(authorList)-1], ", ")
		authorString = authorString + " and " + authorList[len(authorList)-1]
	}

	return authorString
}

func sendWebRequestWithRetry(ctx context.Context, client *http.Client, url string, body webhookBody) error {

	var retryTimer float64

	for done := false; done == false; {

		reqBody, err := json.Marshal(body)

		if err != nil {
			return fmt.Errorf("marshaling body: %w", err)
		}

		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqBody))
		if err != nil {
			return fmt.Errorf("creating web request: %w", err)
		}

		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("User-Agent", "DiscordBot (https://discord.com, 0.0.1)")

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Duration(retryTimer * float64(time.Second))):
		}

		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("making web request: %w", err)
		}

		respBody, err := io.ReadAll(io.LimitReader(resp.Body, responseBodyReadLimit))
		if err != nil {
			resp.Body.Close()
			return fmt.Errorf("reading response body: %w", err)
		}
		resp.Body.Close()

		switch {
		case resp.StatusCode >= 200 && resp.StatusCode < 300:
			done = true

		case resp.StatusCode == 429:
			retryTimer, err = strconv.ParseFloat(resp.Header.Get("Retry-After"), 64)
			if err != nil {
				return fmt.Errorf("reading Retry-After header and converting to float64: %w", err)
			}

			nextRetryTime := time.Now().Add(time.Duration(retryTimer * float64(time.Second)))
			if contextDeadlineTime, ok := ctx.Deadline(); ok {
				if nextRetryTime.After(contextDeadlineTime) {
					return fmt.Errorf("rate limited: discord asked us to wait %f seconds: next retry %s would exceed context deadline %s: exiting", retryTimer, nextRetryTime, contextDeadlineTime)
				}
			}
			slog.Info("rate limited", "Retry-After", retryTimer, "StatusCode", resp.StatusCode)

		default:
			return fmt.Errorf("discord returned status %d: %s", resp.StatusCode, respBody)

		}
	}

	return nil
}

func SendWebhook(ctx context.Context, authorNames []string, content string, updatedBooks []book.Book, url string) error {

	chunks := slices.Chunk(updatedBooks, 10)

	i := 0

	client := http.Client{}

	for chunk := range chunks {

		var embeddedBooks []embed

		for _, updatedBook := range chunk {
			embeddedBook := embed{
				Title:       updatedBook.Title,
				Description: "By " + updatedBook.Author,
				Color:       65280,
				Image: &image{
					URL: updatedBook.ImageURL,
				},
			}
			embeddedBooks = append(embeddedBooks, embeddedBook)
		}

		// if i == 0 {
		// 	var plural string
		// 	if len(updatedBooks) != 1 {
		// 		plural = "s"
		// 	}
		// content = fmt.Sprintf("Search for %s found %d book%s", buildAuthorNames(authorNames), len(updatedBooks), plural)
		// }

		var msg string
		if i == 0 {
			msg = content
		}

		reqBody := webhookBody{
			Content: msg,
			Embeds:  embeddedBooks,
		}

		i++

		err := sendWebRequestWithRetry(ctx, &client, url, reqBody)

		if err != nil {
			return fmt.Errorf("sending web request: %w", err)
		}

		slog.Info("sent chunk", "chunk", i, "chunkSize", len(chunk))
	}

	return nil
}
