package sqlstore

import (
	"database/sql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/iryzzh/practicum-go-shortener/internal/app/store"
	"log"
)

type Store struct {
	db *sql.DB
}

func New(dsn string) (s *Store, err error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	s = &Store{
		db: db,
	}

	if err := s.migrate(); err != nil {
		log.Fatal(err)
	}

	return s, db.Ping()
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) migrate() error {
	driver, err := postgres.WithInstance(s.db, &postgres.Config{})
	if err != nil {
		return err
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file://db/migrations/pg",
		"postgres", driver)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
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
