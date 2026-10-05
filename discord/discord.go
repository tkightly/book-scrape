package discord

import (
	"book-scrape/book"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"
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

func SendWebhook(ctx context.Context, authorNames []string, updatedBooks []book.Book, url string) error {

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

		var content string

		if i == 0 {
			var plural string
			if len(updatedBooks) != 1 {
				plural = "s"
			}
			content = fmt.Sprintf("Search for %s found %d book%s", buildAuthorNames(authorNames), len(updatedBooks), plural)
		} else {
		}

		i++

		reqBody, err := json.Marshal(webhookBody{
			Content: content,
			Embeds:  embeddedBooks,
		})

		if err != nil {
			return fmt.Errorf("marshaling body: %w", err)
		}

		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqBody))
		if err != nil {
			return fmt.Errorf("creating web request: %w", err)
		}

		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("User-Agent", "DiscordBot (https://discord.com, 0.0.1)")

		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("making web request: %w", err)
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			body, err := io.ReadAll(io.LimitReader(resp.Body, responseBodyReadLimit))

			if err != nil {
				resp.Body.Close()
				return fmt.Errorf("reading response body: %w", err)
			}

			resp.Body.Close()
			return fmt.Errorf("discord returned status %d: %s", resp.StatusCode, body)
		}
		resp.Body.Close()

	}

	return nil
}
