package book

import (
	"testing"
)

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
			got := BuildAuthorNames(tc.authorList)
			if tc.wantAuthorString != got {
				t.Errorf("got %s, want %s", got, tc.wantAuthorString)
			}

		})
	}

}
