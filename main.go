package main

import (
	"book-scrape/book"
)

func doDiff(storedBooks, updatedBooks []book.Book) (addedBooks []book.Book, removedBooks []book.Book) {

	diffInto := func(from, against []book.Book) (out []book.Book) {

		againstMap := make(map[string]struct{})
		for _, againstBook := range against {
			againstMap[againstBook.ObjectID] = struct{}{}
		}

		// emit only distinct books
		alreadyEmittedMap := make(map[string]struct{})

		for _, fromBook := range from {
			if _, ok := againstMap[fromBook.ObjectID]; ok {
				continue
			}
			if _, ok := alreadyEmittedMap[fromBook.ObjectID]; ok {
				continue
			}
			out = append(out, fromBook)
			alreadyEmittedMap[fromBook.ObjectID] = struct{}{}
		}

		return out
	}

	removedBooks = diffInto(storedBooks, updatedBooks)

	addedBooks = diffInto(updatedBooks, storedBooks)

	return addedBooks, removedBooks
}

func main() {

}
