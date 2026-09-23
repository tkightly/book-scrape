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

	books := []book.Book{}

	return books, nil
}

func (d *Db) SyncBooks(ctx context.Context, added, removed []book.Book) error {

	return nil
}
