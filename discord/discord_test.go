package discord

import (
	"book-scrape/book"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

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

	tests := []struct {
		name              string
		books             []book.Book
		wantRequests      int
		wantRequestBodies []webhookBody
	}{
		{name: "no books sends nothing", books: nil, wantRequests: 0},
		{
			name: "one book sends single request",
			books: []book.Book{
				{
					ObjectID: "0001",
					Title:    "A book",
					Author:   "Author",
					ImageURL: "https://example.com/image.jpg",
				},
			},
			wantRequests: 1,
			wantRequestBodies: []webhookBody{
				{
					Content: "1 book",
					Embeds: []embed{
						{
							Title:       "A book",
							Description: "By Author",
							Color:       32,
							Image: &image{
								URL: "https://example.com/image.jpg"},
						},
					},
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

			err := SendWebhook(context.Background(), tc.books, server.URL)

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
