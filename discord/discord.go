package discord

import (
	"book-scrape/book"
	"context"
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

// TODO: implement to make TestSendWebhook pass (getPage in algolia has a similar shape)
// TODO: check Discord's max embeds per message - what happens with 25 new books?
func SendWebhook(ctx context.Context, updatedBooks []book.Book, url string) error {

	return nil
}
