package sqlite

import (
	"book-scrape/book"
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
)

type columnInfo struct {
	name string
	typ  string
}

func seedDb(t *testing.T, ctx context.Context, filePath string) (*Db, []book.Book) {

	t.Helper()

	path := filepath.Join(t.TempDir(), filePath)

	db, err := NewDb(path)

	if err != nil {
		t.Fatalf("error: %s", err)
	}

	t.Cleanup(func() {
		err := db.Close()
		if err != nil {
			t.Errorf("error: %s", err)
		}
	})

	_, err = os.Stat(path)

	if err != nil {
		t.Fatalf("error: %s", err)
	}

	expectedBooks := []book.Book{
		{
			ObjectID: "0001",
			Title:    "Test Book 1",
			Author:   "Test Author 1",
			ImageURL: "https://image.com/testImage1.jpg",
			URL:      "https://book.com/book0001.jpg",
			Price:    "£1",
		},
		{
			ObjectID: "0002",
			Title:    "Test Book 2",
			Author:   "Test Author 2",
			ImageURL: "https://image.com/testImage2.jpg",
			URL:      "https://book.com/book2.jpg",
			Price:    "£2",
		},
		{
			ObjectID: "0003",
			Title:    "Test Book 3",
			Author:   "Test Author",
			ImageURL: "https://image.com/testImage3.jpg",
			URL:      "https://book.com/book3.jpg",
			Price:    "£3",
		},
		{
			ObjectID: "0004",
			Title:    "Test Book 4",
			Author:   "Test Author 4",
			ImageURL: "https://image.com/testImage4.jpg",
			URL:      "https://book.com/book4.jpg",
			Price:    "£4",
		},
	}

	stmt, err := db.conn.PrepareContext(ctx, "INSERT INTO books (ObjectID, Title, Author, ImageURL, URL, Price) VALUES (?,?,?,?,?,?) ")

	if err != nil {
		t.Fatalf("error: %s", err)
	}

	defer stmt.Close()

	for _, book := range expectedBooks {

		if _, err := stmt.ExecContext(ctx, book.ObjectID, book.Title, book.Author, book.ImageURL, book.URL, book.Price); err != nil {
			t.Fatalf("error: %s", err)
		}

	}

	return db, expectedBooks

}

func TestNewDb(t *testing.T) {
	path := filepath.Join(t.TempDir(), "db.db")

	ctx := context.Background()

	db, err := NewDb(path)

	if err != nil {
		t.Fatalf("error: %s", err)
	}

	_, err = os.Stat(path)

	if err != nil {
		t.Errorf("error: %s", err)
	}

	rows, err := db.conn.QueryContext(ctx, "SELECT name, type FROM pragma_table_info('books')")

	if err != nil {
		t.Errorf("error: %s", err)
	}

	expectedColumnInfo := [6]columnInfo{
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
		{
			name: "URL",
			typ:  "TEXT",
		},
		{
			name: "Price",
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

	if err := db.Close(); err != nil {
		t.Fatalf("Could not close db: %v", err)
	}
}

func TestClose(t *testing.T) {
	path := filepath.Join(t.TempDir(), "db.db")
	db, err := NewDb(path)

	if err != nil {
		t.Fatalf("could not create database: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Errorf("could not close the db: %v", err)
	}

	_, err = db.ListBooks(context.Background())
	if err == nil {
		t.Errorf("was able to execute query against the db after it should be closed")
	}
}

func TestListBooks(t *testing.T) {

	ctx := context.Background()

	db, expectedBooks := seedDb(t, ctx, "db.db")

	books, err := db.ListBooks(ctx)

	if err != nil {
		t.Fatalf("error: %s", err)
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
			t.Errorf("got %v, want %v", want, got)
		}
	}
}

func TestSyncBooks(t *testing.T) {

	t.Run("insert and delete", func(t *testing.T) {

		ctx := context.Background()

		db, _ := seedDb(t, ctx, "db.db")

		addBooks := []book.Book{
			{
				ObjectID: "0005",
				Title:    "Test Book 5",
				Author:   "Test Author 5",
				ImageURL: "https://image.com/testImage5.jpg",
				URL:      "https://book.com/book5.jpg",
				Price:    "£5",
			},
		}

		removeBooks := []book.Book{
			{
				ObjectID: "0004",
				Title:    "Test Book 4",
				Author:   "Test Author",
				ImageURL: "https://image.com/testImage4.jpg",
				URL:      "https://book.com/book4.jpg",
				Price:    "£4",
			},
		}

		err := db.SyncBooks(ctx, addBooks, removeBooks)

		if err != nil {
			t.Errorf("error: %s", err)
		}

		books, err := db.ListBooks(ctx)

		expectedSyncedBooks := [4]book.Book{
			{
				ObjectID: "0001",
				Title:    "Test Book 1",
				Author:   "Test Author 1",
				ImageURL: "https://image.com/testImage1.jpg",
				URL:      "https://book.com/book0001.jpg",
				Price:    "£1",
			},
			{
				ObjectID: "0002",
				Title:    "Test Book 2",
				Author:   "Test Author 2",
				ImageURL: "https://image.com/testImage2.jpg",
				URL:      "https://book.com/book2.jpg",
				Price:    "£2",
			},
			{
				ObjectID: "0003",
				Title:    "Test Book 3",
				Author:   "Test Author",
				ImageURL: "https://image.com/testImage3.jpg",
				URL:      "https://book.com/book3.jpg",
				Price:    "£3",
			},
			{
				ObjectID: "0005",
				Title:    "Test Book 5",
				Author:   "Test Author 5",
				ImageURL: "https://image.com/testImage5.jpg",
				URL:      "https://book.com/book5.jpg",
				Price:    "£5",
			},
		}

		if err != nil {
			t.Errorf("error: %s", err)
		}

		if len(books) != len(expectedSyncedBooks) {
			t.Errorf("error: retrieved book count does not match: want %v, got %v", len(expectedSyncedBooks), len(books))
		}

		booksMap := make(map[string]book.Book)

		for _, book := range books {

			booksMap[book.ObjectID] = book

		}

		for _, want := range expectedSyncedBooks {

			if got, ok := booksMap[want.ObjectID]; ok != true {
				t.Errorf("Expected data not present: want %v, got %v", want.ObjectID, got.ObjectID)
			} else if want != got {
				t.Errorf("want %v, got %v", want, got)
			}

		}
	})

	t.Run("update via upsert", func(t *testing.T) {

		ctx := context.Background()

		db, _ := seedDb(t, ctx, "db.db")

		addBooks := []book.Book{
			{
				ObjectID: "0004",
				Title:    "Test Book 4 - updated title",
				Author:   "Test Author 4 - new author",
				ImageURL: "https://image.com/testImage4.jpg",
				URL:      "https://book.com/book4.jpg",
				Price:    "£4",
			},
		}

		err := db.SyncBooks(ctx, addBooks, []book.Book{})

		if err != nil {
			t.Errorf("error: %s", err)
		}

		books, err := db.ListBooks(ctx)

		expectedBooks := [4]book.Book{
			{
				ObjectID: "0001",
				Title:    "Test Book 1",
				Author:   "Test Author 1",
				ImageURL: "https://image.com/testImage1.jpg",
				URL:      "https://book.com/book0001.jpg",
				Price:    "£1",
			},
			{
				ObjectID: "0002",
				Title:    "Test Book 2",
				Author:   "Test Author 2",
				ImageURL: "https://image.com/testImage2.jpg",
				URL:      "https://book.com/book2.jpg",
				Price:    "£2",
			},
			{
				ObjectID: "0003",
				Title:    "Test Book 3",
				Author:   "Test Author",
				ImageURL: "https://image.com/testImage3.jpg",
				URL:      "https://book.com/book3.jpg",
				Price:    "£3",
			},
			{
				ObjectID: "0004",
				Title:    "Test Book 4 - updated title",
				Author:   "Test Author 4 - new author",
				ImageURL: "https://image.com/testImage4.jpg",
				URL:      "https://book.com/book4.jpg",
				Price:    "£4",
			},
		}

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
				t.Errorf("got %v, want %v", want, got)
			}

		}
	})

	t.Run("added wins over removed on conflict", func(t *testing.T) {

		ctx := context.Background()

		db, _ := seedDb(t, ctx, "db.db")

		addBooks := []book.Book{
			{
				ObjectID: "0006",
				Title:    "Test Book 6",
				Author:   "Test Author 6",
				ImageURL: "https://image.com/testImage6.jpg",
				URL:      "https://book.com/book6.jpg",
				Price:    "£6",
			},
		}

		removedBooks := []book.Book{
			{
				ObjectID: "0006",
				Title:    "Test Book 6",
				Author:   "Test Author 6",
				ImageURL: "https://image.com/testImage6.jpg",
				URL:      "https://book.com/book6.jpg",
				Price:    "£6",
			},
		}

		err := db.SyncBooks(ctx, addBooks, removedBooks)

		if err != nil {
			t.Errorf("error: %s", err)
		}

		books, err := db.ListBooks(ctx)

		expectedBooks := [5]book.Book{
			{
				ObjectID: "0001",
				Title:    "Test Book 1",
				Author:   "Test Author 1",
				ImageURL: "https://image.com/testImage1.jpg",
				URL:      "https://book.com/book0001.jpg",
				Price:    "£1",
			},
			{
				ObjectID: "0002",
				Title:    "Test Book 2",
				Author:   "Test Author 2",
				ImageURL: "https://image.com/testImage2.jpg",
				URL:      "https://book.com/book2.jpg",
				Price:    "£2",
			},
			{
				ObjectID: "0003",
				Title:    "Test Book 3",
				Author:   "Test Author",
				ImageURL: "https://image.com/testImage3.jpg",
				URL:      "https://book.com/book3.jpg",
				Price:    "£3",
			},
			{
				ObjectID: "0004",
				Title:    "Test Book 4",
				Author:   "Test Author 4",
				ImageURL: "https://image.com/testImage4.jpg",
				URL:      "https://book.com/book4.jpg",
				Price:    "£4",
			},
			{
				ObjectID: "0006",
				Title:    "Test Book 6",
				Author:   "Test Author 6",
				ImageURL: "https://image.com/testImage6.jpg",
				URL:      "https://book.com/book6.jpg",
				Price:    "£6",
			},
		}

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
				t.Errorf("got %v, want %v", want, got)
			}

		}
	})
}

func TestNewDbMigratesOldSchema(t *testing.T) {

	path := filepath.Join(t.TempDir(), "old.db")

	old, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("error: %s", err)
	}
	_, err = old.Exec("CREATE TABLE books(ObjectID TEXT PRIMARY KEY, Title TEXT NOT NULL, Author TEXT NOT NULL, ImageURL TEXT NOT NULL)")
	if err != nil {
		t.Fatalf("error: %s", err)
	}
	_, err = old.Exec("INSERT INTO books VALUES ('1', 'Title 1', 'Author 1', 'https://image.com/1.jpg')")
	if err != nil {
		t.Fatalf("error: %s", err)
	}
	old.Close()

	// opening twice also proves the migration is idempotent
	for range 2 {
		db, err := NewDb(path)
		if err != nil {
			t.Fatalf("NewDb: %s", err)
		}

		books, err := db.ListBooks(context.Background())
		if err != nil {
			t.Fatalf("ListBooks: %s", err)
		}
		if len(books) != 1 || books[0].ObjectID != "1" || books[0].URL != "" || books[0].Price != "" {
			t.Errorf("unexpected books after migration: %+v", books)
		}

		db.Close()
	}
}
