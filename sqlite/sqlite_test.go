package sqlite

import (
	"book-scrape/book"
	"context"
	"os"
	"path/filepath"
	"testing"
)

type columnInfo struct {
	name string
	typ  string
}

func TestNewDb(t *testing.T) {
	path := filepath.Join(t.TempDir(), "db.db")

	ctx := context.Background()

	db, err := NewDb(path)

	if err != nil {
		t.Errorf("error: %s", err)
	}

	_, err = os.Stat(path)

	if err != nil {
		t.Errorf("error: %s", err)
	}

	rows, err := db.conn.QueryContext(ctx, "SELECT name, type FROM pragma_table_info('books')")

	if err != nil {
		t.Errorf("error: %s", err)
	}

	expectedColumnInfo := [4]columnInfo{
		{
			name: "ObjectID",
			typ:  "TEXT",
		},
		{
			name: "Title",
			typ:  "TEXT",
		},
		{
			name: "Author",
			typ:  "TEXT",
		},
		{
			name: "ImageURL",
			typ:  "TEXT",
		},
	}

	defer rows.Close()

	i := 0

	for rows.Next() {

		var columnInfo columnInfo

		if err := rows.Scan(&columnInfo.name, &columnInfo.typ); err != nil {
			t.Errorf("error: %s", err)
		}

		if i > len(expectedColumnInfo)-1 {
			t.Errorf("More columns than expected: got %v, want %v", i, len(expectedColumnInfo)-1)
			break
		}

		if columnInfo.name != expectedColumnInfo[i].name {
			t.Errorf("Unexpected column name: got %v, want %v", columnInfo.name, expectedColumnInfo[i].name)
		}

		if columnInfo.typ != expectedColumnInfo[i].typ {
			t.Errorf("Unexpected column type: got %v, want %v", columnInfo.typ, expectedColumnInfo[i].typ)
		}

		i++

	}

	if i != len(expectedColumnInfo) {
		t.Errorf("Unexpected number of colums: got %v, want %v", i, len(expectedColumnInfo))
	}

	if err := rows.Err(); err != nil {
		t.Errorf("error: %s", err)
	}

}

func TestListBooks(t *testing.T) {
	path := filepath.Join(t.TempDir(), "db.db")

	ctx := context.Background()

	db, err := NewDb(path)

	if err != nil {
		t.Errorf("error: %s", err)
	}

	_, err = os.Stat(path)

	if err != nil {
		t.Errorf("error: %s", err)
	}

	expectedBooks := [4]book.Book{
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

	stmt, err := db.conn.PrepareContext(ctx, "INSERT INTO books (ObjectID, Title, Author, ImageURL) VALUES (?,?,?,?) ")

	if err != nil {
		t.Errorf("error: %s", err)
	}

	defer stmt.Close()

	for _, book := range expectedBooks {

		if _, err := stmt.ExecContext(ctx, book.ObjectID, book.Title, book.Author, book.ImageURL); err != nil {
			t.Errorf("error: %s", err)
		}

	}

	books, err := db.ListBooks(ctx)

	if err != nil {
		t.Errorf("error: %s", err)
	}

	if len(books) != len(expectedBooks) {
		t.Errorf("error: retrieved book count does not match: want %v, got %v", len(expectedBooks), len(books))
	}

	booksMap := make(map[string]book.Book)

	for _, book := range books {

		booksMap[book.ObjectID] = book

	}

	for _, want := range expectedBooks {

		if got, ok := booksMap[want.ObjectID]; ok != true {
			t.Errorf("Expected data not present: want %v, got %v", want.ObjectID, got.ObjectID)
		} else if want != got {
			t.Errorf("Wanted row dosn't match got row")
		}

	}

}
