package postgres

import (
	"context"
	"errors"
	"fmt"
	"log"
	"restapiserv/internal/storage"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	db *pgxpool.Pool
}

func New(storagePath string) (*Storage, error) {
	const op = "storage.postgres.New"

	db, err := pgxpool.New(context.Background(), storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	_, err = db.Exec(context.Background(), `
	CREATE TABLE IF NOT EXISTS url(
		id SERIAL PRIMARY KEY, 
		alias TEXT NOT NULL UNIQUE,
		url TEXT NOT NULL);
	CREATE INDEX IF NOT EXISTS idx_alias ON url(alias);
	`)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveURL(urlToSave string, alias string) (int64, error) {
	const fn = "storage.postgres.SaveURL"

	// Умная вставка с обработкой дубликатов
	var id int64
	err := s.db.QueryRow(
		context.Background(),
		`INSERT INTO url(alias, url) VALUES($1, $2) 
         ON CONFLICT (alias) DO NOTHING
         RETURNING id`,
		alias,
		urlToSave,
	).Scan(&id)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, storage.ErrURLExists
		}
		log.Printf("%s: failed to save URL: %v\nURL: %s\nAlias: %s",
			fn, err, urlToSave, alias) // Детальное логирование
		return 0, fmt.Errorf("%s: %w", fn, err)
	}
	return id, nil
}

func (s *Storage) GetURL(alias string) (string, error) {
	const fn = "storage.postgres.GetURL"

	var url string
	err := s.db.QueryRow(context.Background(),
		"SELECT url FROM url WHERE alias = $1", alias).Scan(&url) // .Scan(&url) необходим для извлечения данных из результата SQL-запроса.

	if err != nil {
		log.Printf("Database error: %v", err) // Debug log
		if errors.Is(err, pgx.ErrNoRows) {
			return "", storage.ErrURLNotFound
		}
		return "", fmt.Errorf("%s: execute statement %w", fn, err)
	}

	return url, nil
}

func (s *Storage) DeleteURL(alias string) (int64, error) {
	const fn = "storage.postgres.DeleteURL"

	result, err := s.db.Exec(context.Background(),
		"DELETE FROM url WHERE alias = $1", alias)
	if err != nil {
		return 0, fmt.Errorf("%s: execute statement %w", fn, err)
	}

	rowsAffected := result.RowsAffected() // Считаем сколько удалили

	return rowsAffected, nil
}

func (s *Storage) UpdateURL(alias string, newURL string) (int64, error) {
	const fn = "storage.postgres.UpdateURL"

	result, err := s.db.Exec(context.Background(),
		"UPDATE url SET url = $1 WHERE alias = $2", newURL, alias)

	if err != nil {
		return 0, fmt.Errorf("%s: execute statement %w", fn, err)
	}

	rowsAffected := result.RowsAffected()
	return rowsAffected, nil
}

func (s *Storage) Close() {
	s.db.Close()
}
