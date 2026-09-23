package sqlite

import (
	"book-scrape/book"
	"context"
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

type Db struct {
	conn *sql.DB
}

func NewDb(path string) (*Db, error) {

	conn, err := sql.Open("sqlite", path)

	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	db := Db{
		conn: conn,
	}

	_, err = db.conn.Exec("CREATE TABLE IF NOT EXISTS books(ObjectID TEXT PRIMARY KEY, Title TEXT NOT NULL, Author TEXT NOT NULL, ImageURL TEXT NOT NULL)")

	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return &db, nil
}

func (d *Db) ListBooks(ctx context.Context) ([]book.Book, error) {

	query := "SELECT ObjectID, Title, Author, ImageURL FROM books"

	rows, err := d.conn.QueryContext(ctx, query)

	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	defer rows.Close()

	var books []book.Book

	for rows.Next() {

		var book book.Book

		if err := rows.Scan(&book.ObjectID, &book.Title, &book.Author, &book.ImageURL); err != nil {
			return nil, fmt.Errorf("%w", err)

		}

		books = append(books, book)

	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return books, nil
}

func (d *Db) SyncBooks(ctx context.Context, added, removed []book.Book) error {

	return nil
}
