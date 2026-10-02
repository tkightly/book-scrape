package main

import (
	"book-scrape/book"
	"testing"
)

func TestDoDiff(t *testing.T) {

	storedBooks := []book.Book{
		{
			ObjectID: "0001",
			Title:    "Test Book 1",
			Author:   "Test Author 1",
			ImageURL: "https://image.com/testImage1.jpg",
		},
		{
			ObjectID: "0002",
			Title:    "Test Book 2",
			Author:   "Test Author 2",
			ImageURL: "https://image.com/testImage2.jpg",
		},
		{
			ObjectID: "0003",
			Title:    "Test Book 3",
			Author:   "Test Author",
			ImageURL: "https://image.com/testImage3.jpg",
		},
		{
			ObjectID: "0004",
			Title:    "Test Book 4",
			Author:   "Test Author 4",
			ImageURL: "https://image.com/testImage4.jpg",
		},
	}

	updatedBooks := []book.Book{
		{
			ObjectID: "0001",
			Title:    "Test Book 1",
			Author:   "Test Author 1",
			ImageURL: "https://image.com/testImage1.jpg",
		},
		{
			ObjectID: "0002",
			Title:    "Test Book 2",
			Author:   "Test Author 2",
			ImageURL: "https://image.com/testImage2.jpg",
		},
		{
			ObjectID: "0003",
			Title:    "Test Book 3",
			Author:   "Test Author",
			ImageURL: "https://image.com/testImage3.jpg",
		},
		{
			ObjectID: "0005",
			Title:    "Test Book 5",
			Author:   "Test Author 5",
			ImageURL: "https://image.com/testImage5.jpg",
		},
	}

	expectedAddedBooks := []book.Book{
		{
			ObjectID: "0005",
			Title:    "Test Book 5",
			Author:   "Test Author 5",
			ImageURL: "https://image.com/testImage5.jpg",
		},
	}

	expectedRemovedBooks := []book.Book{
		{
			ObjectID: "0004",
			Title:    "Test Book 4",
			Author:   "Test Author 4",
			ImageURL: "https://image.com/testImage4.jpg",
		},
	}

	expectedAddedBooksMap := make(map[string]book.Book)

	for _, expectedAddedBook := range expectedAddedBooks {
		expectedAddedBooksMap[expectedAddedBook.ObjectID] = expectedAddedBook
	}

	expectedRemovedBooksMap := make(map[string]book.Book)

	for _, expectedRemovedBook := range expectedRemovedBooks {
		expectedRemovedBooksMap[expectedRemovedBook.ObjectID] = expectedRemovedBook
	}

	addedBooks, removedBooks := doDiff(storedBooks, updatedBooks)

	addedBooksMap := make(map[string]book.Book)

	for _, addedBook := range addedBooks {
		addedBooksMap[addedBook.ObjectID] = addedBook
	}

	removedBooksMap := make(map[string]book.Book)

	for _, removedBook := range removedBooks {
		removedBooksMap[removedBook.ObjectID] = removedBook
	}

	t.Run("quantity of books should match expected", func(t *testing.T) {
		if len(addedBooks) != len(expectedAddedBooks) {
			t.Errorf("Unexpected quantity of books to be added; got %v, want %v", len(addedBooks), len(expectedAddedBooks))
		}

		if len(removedBooks) != len(expectedRemovedBooks) {
			t.Errorf("Unexpected quantity of books to be removed; got %v, want %v", len(removedBooks), len(expectedRemovedBooks))
		}
	})

	t.Run("expected books should be in returned list", func(t *testing.T) {
		for _, want := range expectedAddedBooks {
			if _, ok := addedBooksMap[want.ObjectID]; !ok {
				t.Errorf("%s book not in the returned list to be added", want.ObjectID)
			}
		}

		for _, want := range expectedRemovedBooks {
			if _, ok := removedBooksMap[want.ObjectID]; !ok {
				t.Errorf("%s book not in the returned list to be removed", want.ObjectID)
			}
		}
	})

	t.Run("unexpected books should not be returned", func(t *testing.T) {
		for _, got := range addedBooks {
			if _, ok := expectedAddedBooksMap[got.ObjectID]; !ok {
				t.Errorf("%s book unexpectedly in the returned list to be added", got.ObjectID)
			}
		}

		for _, got := range removedBooks {
			if _, ok := expectedRemovedBooksMap[got.ObjectID]; !ok {
				t.Errorf("%s book unexpectedly in the returned list to be removed", got.ObjectID)
			}
		}
	})

}
