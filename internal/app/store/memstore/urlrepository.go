package memstore

import (
	"fmt"
	"github.com/iryzzh/practicum-go-shortener/internal/app/model"
)

type URLRepository struct {
	store *Store
}

func (r *URLRepository) Create(u *model.URL) error {
	if err := u.Validate(); err != nil {
		return err
	}

	r.store.Lock()
	defer r.store.Unlock()
	if err := u.Validate(); err != nil {
		return err
	}

	u.ID = r.store.nextID + 1

	r.store.urls[r.store.nextID] = u
	r.store.nextID++

	return nil
}

func (r *URLRepository) FindByID(id string) (*model.URL, error) {
	for _, v := range r.store.urls {
		if id == v.URLShort {
			return v, nil
		}
	}
	return &model.URL{}, fmt.Errorf("record not found")
}

func (r *URLRepository) IncrementStats(i int) {
	r.store.Lock()
	defer r.store.Unlock()
	r.store.urls[i].Visited = true
	r.store.urls[i].Count++
}
