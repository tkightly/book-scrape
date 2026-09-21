package algolia

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const indexName string = "shopify_products"

// const algoliaUrl string = "https://algolia.worldofbooks.com/1/indexes/*/queries"

type Response struct {
	Results []SearchResult `json:"results"`
}

type SearchResult struct {
	Hits  []Hit
	Page  int `json:"page"`
	Pages int `json:"nbPages"`
}

type Hit struct {
	ObjectID  string `json:"objectID"`
	LongTitle string `json:"longTitle"`
	Author    string `json:"author"`
	ImageURL  string `json:"imageURL"`
}

type searchRequest struct {
	Requests []Request `json:"requests"`
}

type Request struct {
	IndexName string `json:"indexName"`
	Filters   string `json:"filters"`
	Page      int    `json:"page"`
}

func getPage(ctx context.Context, url string, filter string, page int) (SearchResult, error) {

	reqBody, err := json.Marshal(searchRequest{
		Requests: []Request{
			{
				IndexName: indexName,
				Filters:   filter,
				Page:      page,
			},
		},
	})

	if err != nil {
		return SearchResult{}, fmt.Errorf("%w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqBody))

	if err != nil {
		return SearchResult{}, fmt.Errorf("%w", err)
	}

	req.Header.Set("x-algolia-api-key", "proxy")
	req.Header.Set("x-algolia-application-id", "AR33G9NJGJ")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "text/plain") // TODO see if application.json would work here too
	req.Header.Set("Origin", "https://www.worldofbooks.com")

	client := http.Client{}

	resp, err := client.Do(req)

	if err != nil {
		return SearchResult{}, fmt.Errorf("%w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return SearchResult{}, fmt.Errorf("algolia returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return SearchResult{}, fmt.Errorf("%w", err)
	}

	var response Response

	err = json.Unmarshal(body, &response)

	if err != nil {
		return SearchResult{}, fmt.Errorf("%w", err)
	}

	return response.Results[0], nil

}

// func Search(ctx context.Context, author string) ([]book.Book, error) {
//
// 	return book, err
// }
