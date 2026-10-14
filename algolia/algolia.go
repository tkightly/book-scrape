package algolia

import (
	"book-scrape/book"
	"bytes"
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const indexName string = "shopify_products"
const responseBodyReadLimit int64 = 1024000 // 1Mb

var algoliaUrl string = "https://algolia.worldofbooks.com/1/indexes/*/queries"

type Response struct {
	Results []SearchResult `json:"results"`
}

type SearchResult struct {
	Hits  []Hit
	Page  int `json:"page"`
	Pages int `json:"nbPages"`
}

type Hit struct {
	ObjectID           string  `json:"objectID"`
	LongTitle          string  `json:"longTitle"`
	Author             string  `json:"author"`
	ImageURL           string  `json:"imageURL"`
	BestConditionPrice float64 `json:"bestConditionPrice"`
	FromPrice          float64 `json:"fromPrice"`
	ProductHandle      string  `json:"productHandle"`
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
		return SearchResult{}, fmt.Errorf("marshaling searchRequest: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqBody))

	if err != nil {
		return SearchResult{}, fmt.Errorf("creating http request: %w", err)
	}

	req.Header.Set("x-algolia-api-key", "proxy")
	req.Header.Set("x-algolia-application-id", "AR33G9NJGJ")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://www.worldofbooks.com")

	client := http.Client{}

	resp, err := client.Do(req)

	if err != nil {
		return SearchResult{}, fmt.Errorf("making web request: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return SearchResult{}, fmt.Errorf("algolia returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, responseBodyReadLimit))

	if err != nil {
		return SearchResult{}, fmt.Errorf("reading response body: %w", err)
	}

	var response Response

	err = json.Unmarshal(body, &response)

	if err != nil {
		return SearchResult{}, fmt.Errorf("unmarshaling response body (%dKb): %w", (responseBodyReadLimit / 1000), err)
	}

	if len(response.Results) > 0 {
		return response.Results[0], nil
	} else {
		return SearchResult{}, nil
	}

}

func Search(ctx context.Context, author string) ([]book.Book, error) {

	filter := fmt.Sprintf(`fromPrice > 0 AND author:%q`, author)

	page := 0
	var books []book.Book
	for {
		results, err := getPage(ctx, algoliaUrl, filter, page)

		if err != nil {
			return []book.Book{}, fmt.Errorf("getting results page %d for author %q: %w", page, author, err)
		}

		for _, hit := range results.Hits {
			hitBook := book.Book{
				ObjectID: hit.ObjectID,
				Title:    hit.LongTitle,
				Author:   hit.Author,
				ImageURL: hit.ImageURL,
				URL:      fmt.Sprintf("https://www.worldofbooks.com/en-gb/products/%s", hit.ProductHandle),
				Price:    fmt.Sprintf("£%.2f", cmp.Or(hit.BestConditionPrice, hit.FromPrice)),
			}

			books = append(books, hitBook)
		}

		page++

		if page >= results.Pages {
			break
		}

	}

	return books, nil
}
