package storage

import (
	"database/sql"

	_ "github.com/lib/pq"
)

func NewPostgres(url string) (*sql.DB, error) {
	return sql.Open("postgres", url)
}
