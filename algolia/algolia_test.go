package algolia

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetPage(t *testing.T) {

	filter := `fromPrice > 0 AND author:"Gene Wolfe"`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		body, err := io.ReadAll(r.Body)

		if err != nil {
			t.Errorf("error: %s", err)
		}

		var jsonBody searchRequest

		err = json.Unmarshal(body, &jsonBody)

		if err != nil {
			t.Errorf("error: %s", err)
		}

		got := jsonBody.Requests[0].Filters

		if got != filter {
			t.Errorf("got %v, want %v", got, filter)
		}

		resp := Response{Results: []SearchResult{{Hits: []Hit{{ObjectID: "test-id"}}, Page: 0, Pages: 1}}}

		respJson, err := json.Marshal(resp)

		if err != nil {
			t.Errorf("error: %s", err)
		}

		_, err = w.Write(respJson)
		if err != nil {
			t.Errorf("error: %s", err)
		}

	}))

	defer server.Close()

	ctx := context.Background()

	url := server.URL

	page := 1

	searchResult, err := getPage(ctx, url, filter, page)

	if err != nil {
		t.Errorf("error: %s", err)
	}

	if len(searchResult.Hits) < 1 {
		t.Errorf("No search results found")
	} else if searchResult.Hits[0].ObjectID != "test-id" {
		t.Errorf("Got %v, want %v", searchResult.Hits[0].ObjectID, "test-id")
	}

}
