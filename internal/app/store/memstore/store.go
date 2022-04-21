package memstore

import (
	"github.com/iryzzh/practicum-go-shortener/internal/app/model"
	"github.com/iryzzh/practicum-go-shortener/internal/app/store"
	"sync"
)

type Store struct {
	sync.Mutex

	urls   map[int]*model.URL
	nextID int
}

func New() *Store {
	return &Store{
		urls:   make(map[int]*model.URL),
		nextID: 0,
	}
}

func (s *Store) URL() store.URLRepository {
	return &URLRepository{store: s}
}
