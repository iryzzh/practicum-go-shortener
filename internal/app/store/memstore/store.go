package memstore

import (
	"github.com/iryzzh/practicum-go-shortener/internal/app/model"
	"github.com/iryzzh/practicum-go-shortener/internal/app/store"
	"sync"
)

type Store struct {
	sync.Mutex

	urls       map[int]*model.URL
	users      map[int]*model.User
	urlNextID  int
	userNextID int
}

func (s *Store) Close() error {
	return nil
}

func New() *Store {
	return &Store{
		urls:       make(map[int]*model.URL),
		users:      make(map[int]*model.User),
		urlNextID:  0,
		userNextID: 0,
	}
}

func (s *Store) URL() store.URLRepository {
	return &URLRepository{store: s}
}

func (s *Store) User() store.UserRepository {
	return &UserRepository{store: s}
}

func (s *Store) Ping() error {
	return nil
}
