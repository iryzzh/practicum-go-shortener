package memstore

import (
	"github.com/iryzzh/practicum-go-shortener/internal/app/model"
	"github.com/iryzzh/practicum-go-shortener/internal/app/store"
)

type URLRepository struct {
	store *Store
}

func (r *URLRepository) Create(u *model.URL) error {
	r.store.Lock()
	defer r.store.Unlock()
	if err := u.Validate(); err != nil {
		return err
	}

	u.ID = r.store.urlNextID + 1

	r.store.urls[r.store.urlNextID] = u
	r.store.urlNextID++

	return nil
}

func (r *URLRepository) Exists(u *model.URL) bool {
	for _, v := range r.store.urls {
		if v.URLShort == u.URLShort {
			return true
		}
	}

	return false
}

func (r *URLRepository) FindByID(id int) (*model.URL, error) {
	for _, v := range r.store.urls {
		if id == v.ID {
			return v, nil
		}
	}

	return nil, store.ErrRecordNotFound
}

func (r *URLRepository) FindByUUID(uuid string) (*model.URL, error) {
	for _, v := range r.store.urls {
		if v.URLShort == uuid {
			return v, nil
		}
	}

	return nil, store.ErrRecordNotFound
}

func (r *URLRepository) IncrementStats(i int) {
	r.store.Lock()
	defer r.store.Unlock()
	r.store.urls[i].Visited = true
	r.store.urls[i].Count++
}
