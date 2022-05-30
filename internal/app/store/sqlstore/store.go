package sqlstore

import (
	"database/sql"
	"github.com/iryzzh/practicum-go-shortener/internal/app/store"
)

type Store struct {
	db *sql.DB
}

func New(db *sql.DB) *Store {
	return &Store{
		db: db,
	}
}

func (s *Store) URL() store.URLRepository {
	return &URLRepository{store: s}
}

func (s *Store) User() store.UserRepository {
	return &UserRepository{store: s}
}

func (s *Store) Ping() error {
	return s.db.Ping()
}
