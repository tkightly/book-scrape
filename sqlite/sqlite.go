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

func (d *Db) Close() error {
	return d.conn.Close()
}

func NewDb(path string) (*Db, error) {

	conn, err := sql.Open("sqlite", path)

	if err != nil {
		return nil, fmt.Errorf("opening db at %q: %w", path, err)
	}

	db := Db{
		conn: conn,
	}

	_, err = db.conn.Exec("CREATE TABLE IF NOT EXISTS books(ObjectID TEXT PRIMARY KEY, Title TEXT NOT NULL, Author TEXT NOT NULL, ImageURL TEXT NOT NULL)")

	if err != nil {

		// the db.conn.Exec error is the error we really care about
		_ = conn.Close()
		return nil, fmt.Errorf("creating table: %w", err)
	}

	return &db, nil
}

func (d *Db) ListBooks(ctx context.Context) ([]book.Book, error) {

	query := "SELECT ObjectID, Title, Author, ImageURL FROM books"

	rows, err := d.conn.QueryContext(ctx, query)

	if err != nil {
		return nil, fmt.Errorf("executing query %q: %w", query, err)
	}

	defer rows.Close()

	var books []book.Book

	i := 0

	for rows.Next() {

		var book book.Book
		i++

		if err := rows.Scan(&book.ObjectID, &book.Title, &book.Author, &book.ImageURL); err != nil {
			return nil, fmt.Errorf("scanning row %d: %w", i, err)

		}

		books = append(books, book)

	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("looping rows: %w", err)
	}

	return books, nil
}

func (d *Db) SyncBooks(ctx context.Context, added, removed []book.Book) error {

	// added should win over removed
	addedMap := make(map[string]struct{})

	for _, addedBook := range added {
		addedMap[addedBook.ObjectID] = struct{}{}
	}

	tx, err := d.conn.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return fmt.Errorf("beginning db transaction: %w", err)
	}
	defer tx.Rollback()

	addedStmt, err := tx.PrepareContext(ctx, "INSERT INTO books (ObjectID, Title, Author, ImageURL) VALUES (?, ?, ?, ?) ON CONFLICT(ObjectID) DO UPDATE SET Title = excluded.Title, Author = excluded.Author, ImageURL = excluded.ImageURL")
	if err != nil {
		return fmt.Errorf("preparing INSERT statement: %w", err)
	}

	defer addedStmt.Close()

	for _, addedBook := range added {
		if _, err := addedStmt.ExecContext(ctx, addedBook.ObjectID, addedBook.Title, addedBook.Author, addedBook.ImageURL); err != nil {
			return fmt.Errorf("executing INSERT statement for added book id %q, title %q: %w", addedBook.ObjectID, addedBook.Title, err)
		}
	}

	removedStmt, err := tx.PrepareContext(ctx, "DELETE FROM books WHERE ObjectID == ?")
	if err != nil {
		return fmt.Errorf("preparing DELETE statement: %w", err)
	}

	defer removedStmt.Close()

	for _, removedBook := range removed {
		if _, ok := addedMap[removedBook.ObjectID]; !ok {
			if _, err := removedStmt.ExecContext(ctx, removedBook.ObjectID); err != nil {
				return fmt.Errorf("executing DELETE statement for removed book id %q, title %q: %w", removedBook.ObjectID, removedBook.Title, err)
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}
	return nil
}
