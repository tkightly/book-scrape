package main

import (
	"book-scrape/book"
	"slices"
	"strings"
	"testing"
)

func makeBooks(ids ...string) (books []book.Book) {
	for _, id := range ids {
		books = append(books, book.Book{ObjectID: id, Title: "Title " + id, Author: "Author " + id, ImageURL: "https://image.com/testImage" + id + ".jpg"})
	}

	return books
}

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
			stored:      makeBooks("0001", "0004"),
			updated:     makeBooks("0001", "0005"),
			wantAdded:   makeBooks("0005"),
			wantRemoved: makeBooks("0004"),
		},
		{
			name:        "empty store, empty update",
			stored:      makeBooks(),
			updated:     makeBooks(),
			wantAdded:   makeBooks(),
			wantRemoved: makeBooks(),
		},
		{
			name:        "empty store, many added",
			stored:      makeBooks(),
			updated:     makeBooks("0001", "0005"),
			wantAdded:   makeBooks("0001", "0005"),
			wantRemoved: makeBooks(),
		},
		{
			name:        "some in store, empty added",
			stored:      makeBooks("0001", "0005"),
			updated:     makeBooks(),
			wantAdded:   makeBooks(),
			wantRemoved: makeBooks("0001", "0005"),
		},
		{
			name:        "no change",
			stored:      makeBooks("0001", "0005"),
			updated:     makeBooks("0001", "0005"),
			wantAdded:   makeBooks(),
			wantRemoved: makeBooks(),
		},
		{
			name:        "duplicate ObjectID in updated but in stored",
			stored:      makeBooks("0001", "0005"),
			updated:     makeBooks("0001", "0001", "0005"),
			wantAdded:   makeBooks(),
			wantRemoved: makeBooks(),
		},
		{
			name:        "duplicate ObjectID in stored",
			stored:      makeBooks("0001", "0001", "0005"),
			updated:     makeBooks("0001", "0005"),
			wantAdded:   makeBooks(),
			wantRemoved: makeBooks(),
		},
		{
			name:        "duplicate ObjectID in both",
			stored:      makeBooks("0001", "0001", "0005"),
			updated:     makeBooks("0001", "0001", "0005"),
			wantAdded:   makeBooks(),
			wantRemoved: makeBooks(),
		},
		{
			name:        "duplicate ObjectID in updated, not in stored",
			stored:      makeBooks("0005"),
			updated:     makeBooks("0001", "0001", "0005"),
			wantAdded:   makeBooks("0001"),
			wantRemoved: makeBooks(),
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
