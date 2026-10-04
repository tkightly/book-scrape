package book

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
