package book

import (
	"strings"
)

type Book struct {
	ObjectID string
	Title    string
	Author   string
	ImageURL string
}

func MakeBooks(ids ...string) (books []Book) {
	for _, id := range ids {
		books = append(books, Book{ObjectID: id, Title: "Title " + id, Author: "Author " + id, ImageURL: "https://image.com/testImage" + id + ".jpg"})
	}

	return books
}

func BuildAuthorNames(authorList []string) (authorString string) {

	switch len(authorList) {
	case 0:
		return ""
	case 1:
		return authorList[0]
	default:
		authorString = strings.Join(authorList[:len(authorList)-1], ", ")
		authorString = authorString + " and " + authorList[len(authorList)-1]
	}

	return authorString
}
