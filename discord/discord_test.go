package discord

import (
	"book-scrape/book"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func makeEmbeds(books ...book.Book) (embeds []embed) {
	for _, embeddedBook := range books {
		embeds = append(embeds, embed{
			Title:       embeddedBook.Title,
			Description: fmt.Sprintf("By %s", embeddedBook.Author),
			Color:       65280,
			Image: &image{
				URL: embeddedBook.ImageURL,
			},
		},
		)
	}

	return embeds
}

func TestBuildAuthorNames(t *testing.T) {

	tests := []struct {
		name             string
		authorList       []string
		wantAuthorString string
	}{
		{
			name:             "no authors",
			authorList:       []string{},
			wantAuthorString: "",
		},
		{
			name:             "one author",
			authorList:       []string{"Gene Wolfe"},
			wantAuthorString: "Gene Wolfe",
		},
		{
			name:             "two authors",
			authorList:       []string{"Gene Wolfe", "Iain Banks"},
			wantAuthorString: "Gene Wolfe and Iain Banks",
		},
		{
			name:             "three authors",
			authorList:       []string{"Gene Wolfe", "Ursula K. Le Guin", "Iain Banks"},
			wantAuthorString: "Gene Wolfe, Ursula K. Le Guin and Iain Banks",
		},
		{
			name:             "four authors",
			authorList:       []string{"Gene Wolfe", "Ursula K. Le Guin", "Iain Banks", "Bob"},
			wantAuthorString: "Gene Wolfe, Ursula K. Le Guin, Iain Banks and Bob",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := buildAuthorNames(tc.authorList)
			if tc.wantAuthorString != got {
				t.Errorf("got %s, want %s", got, tc.wantAuthorString)
			}

		})
	}

}

func TestWebhookBodyJSON(t *testing.T) {
	tests := []struct {
		name string
		body webhookBody
		want string
	}{
		{
			name: "content and one embed with an image",
			body: webhookBody{
				Content: "1 book found",
				Embeds: []embed{
					{Title: "Peace", Description: "by Gene Wolfe", Color: 65280, Image: &image{URL: "https://example.com/peace.jpg"}},
				},
			},
			want: `{"content":"1 book found","embeds":[{"title":"Peace","description":"by Gene Wolfe","color":65280,"image":{"url":"https://example.com/peace.jpg"}}]}`,
		},
		{
			name: "no content, no description, no image",
			body: webhookBody{
				Embeds: []embed{
					{Title: "Peace", Color: 65280},
				},
			},
			want: `{"embeds":[{"title":"Peace","color":65280}]}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			jsonBody, err := json.Marshal(tc.body)
			if err != nil {
				t.Fatal(err)
			}

			var got any
			var want any

			err = json.Unmarshal(jsonBody, &got)
			if err != nil {
				t.Fatal(err)
			}

			wantBytes := []byte(tc.want)
			err = json.Unmarshal(wantBytes, &want)
			if err != nil {
				t.Fatal(err)
			}

			if ok := reflect.DeepEqual(want, got); !ok {
				t.Errorf("want %v, got %v", want, got)
			}
		})
	}
}

type recordedRequest struct {
	method      string
	contentType string
	userAgent   string
	body        string
}

func TestSendWebhook(t *testing.T) {

	books := book.MakeBooks("1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11")
	embeds := makeEmbeds(books...)

	tests := []struct {
		name              string
		authors           []string
		books             []book.Book
		wantRequests      int
		wantRequestBodies []webhookBody
	}{
		{
			name:         "no books sends nothing",
			authors:      []string{},
			books:        nil,
			wantRequests: 0,
		},
		{
			name:         "one book sends single request",
			authors:      []string{"Author 1"},
			books:        books[0:1],
			wantRequests: 1,
			wantRequestBodies: []webhookBody{
				{
					Content: "Search for Author 1 found 1 book",
					Embeds:  embeds[0:1],
				},
			},
		},
		{
			name:         "eleven books sends two requests",
			authors:      []string{"Author 1", "Author 2", "Author 3"},
			books:        books,
			wantRequests: 2,
			wantRequestBodies: []webhookBody{
				{
					Content: "Search for Author 1, Author 2 and Author 3 found 11 books",
					Embeds:  embeds[0:10],
				},
				{
					Content: "",
					Embeds:  embeds[10:11],
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var got []recordedRequest
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, err := io.ReadAll(r.Body)
				if err != nil {
					t.Errorf("reading request body: %v", err)
				}

				//record before responding
				got = append(got, recordedRequest{
					method:      r.Method,
					contentType: r.Header.Get("Content-Type"),
					userAgent:   r.Header.Get("User-Agent"),
					body:        string(body),
				})

				w.WriteHeader(http.StatusNoContent)
			}))
			defer server.Close()

			err := SendWebhook(context.Background(), tc.authors, tc.books, server.URL)

			if err != nil {
				t.Fatalf("SendWebhook returned error: %v", err)
			}

			if len(got) != tc.wantRequests {
				t.Errorf("got %d requests, want %d", len(got), tc.wantRequests)
			}

			for i, gotRequest := range got {
				if gotRequest.method != "POST" {
					t.Errorf("method: got %s, want %s", gotRequest.method, "POST")
				}

				if gotRequest.contentType != "application/json" {
					t.Errorf("Content-Type: got %s, want %s", gotRequest.contentType, "application/json")
				}

				if gotRequest.userAgent != "DiscordBot (https://discord.com, 0.0.1)" {
					t.Errorf("User-Agent: got %s, want %s", gotRequest.userAgent, "DiscordBot (https://discord.com, 0.0.1)")
				}

				if i >= len(tc.wantRequestBodies) {
					t.Errorf("got %d requests, want %d", len(got), len(tc.wantRequestBodies))
					continue
				}

				var gotBody webhookBody

				err := json.Unmarshal([]byte(gotRequest.body), &gotBody)

				if err != nil {
					t.Errorf("could not unmarshal request body: %v", err)
				}

				if ok := reflect.DeepEqual(gotBody, tc.wantRequestBodies[i]); !ok {
					t.Errorf("body: got %+v, want %+v", gotBody, tc.wantRequestBodies[i])
				}
			}
		})

	}

}
