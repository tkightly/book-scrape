package main

import (
	"book-scrape/book"
	"slices"
	"strings"
	"testing"
)

func getIds(books []book.Book) (ids []string) {
	for _, thisBook := range books {
		ids = append(ids, thisBook.ObjectID)
	}

	return ids
}

func assertSameBooks(t *testing.T, label string, got, want []book.Book) {

	t.Helper()

	got = slices.Clone(got)
	want = slices.Clone(want)

	slices.SortFunc(got, func(a, b book.Book) int {
		return strings.Compare(a.ObjectID, b.ObjectID)
	})

	slices.SortFunc(want, func(a, b book.Book) int {
		return strings.Compare(a.ObjectID, b.ObjectID)
	})

	if equal := slices.Equal(got, want); !equal {

		t.Errorf("want %s %s, got %s", getIds(want), label, getIds(got))
	}
}

func TestDoDiff(t *testing.T) {
	tests := []struct {
		name        string
		stored      []book.Book
		updated     []book.Book
		wantAdded   []book.Book
		wantRemoved []book.Book
	}{
		{
			name:        "one added, one removed, one unchanged",
			stored:      book.MakeBooks("0001", "0004"),
			updated:     book.MakeBooks("0001", "0005"),
			wantAdded:   book.MakeBooks("0005"),
			wantRemoved: book.MakeBooks("0004"),
		},
		{
			name:        "empty store, empty update",
			stored:      book.MakeBooks(),
			updated:     book.MakeBooks(),
			wantAdded:   book.MakeBooks(),
			wantRemoved: book.MakeBooks(),
		},
		{
			name:        "empty store, many added",
			stored:      book.MakeBooks(),
			updated:     book.MakeBooks("0001", "0005"),
			wantAdded:   book.MakeBooks("0001", "0005"),
			wantRemoved: book.MakeBooks(),
		},
		{
			name:        "some in store, empty added",
			stored:      book.MakeBooks("0001", "0005"),
			updated:     book.MakeBooks(),
			wantAdded:   book.MakeBooks(),
			wantRemoved: book.MakeBooks("0001", "0005"),
		},
		{
			name:        "no change",
			stored:      book.MakeBooks("0001", "0005"),
			updated:     book.MakeBooks("0001", "0005"),
			wantAdded:   book.MakeBooks(),
			wantRemoved: book.MakeBooks(),
		},
		{
			name:        "duplicate ObjectID in updated but in stored",
			stored:      book.MakeBooks("0001", "0005"),
			updated:     book.MakeBooks("0001", "0001", "0005"),
			wantAdded:   book.MakeBooks(),
			wantRemoved: book.MakeBooks(),
		},
		{
			name:        "duplicate ObjectID in stored",
			stored:      book.MakeBooks("0001", "0001", "0005"),
			updated:     book.MakeBooks("0001", "0005"),
			wantAdded:   book.MakeBooks(),
			wantRemoved: book.MakeBooks(),
		},
		{
			name:        "duplicate ObjectID in both",
			stored:      book.MakeBooks("0001", "0001", "0005"),
			updated:     book.MakeBooks("0001", "0001", "0005"),
			wantAdded:   book.MakeBooks(),
			wantRemoved: book.MakeBooks(),
		},
		{
			name:        "duplicate ObjectID in updated, not in stored",
			stored:      book.MakeBooks("0005"),
			updated:     book.MakeBooks("0001", "0001", "0005"),
			wantAdded:   book.MakeBooks("0001"),
			wantRemoved: book.MakeBooks(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			added, removed := doDiff(tc.stored, tc.updated)

			assertSameBooks(t, "added", added, tc.wantAdded)
			assertSameBooks(t, "removed", removed, tc.wantRemoved)

		})
	}
}
