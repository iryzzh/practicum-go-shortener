package memstore

import (
	"fmt"
	"github.com/iryzzh/practicum-go-shortener/internal/app/model"
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

	url := model.URL{
		URLShort:  u.URLShort,
		URLOrigin: u.URLOrigin,
	}

	r.store.urls[r.store.nextID] = url
	r.store.nextID++

	return nil
}

func (r *URLRepository) FindByID(id string) (model.URL, error) {
	for _, v := range r.store.urls {
		if id == v.URLShort {
			return v, nil
		}
	}
	return model.URL{}, fmt.Errorf("record not found")
}

func (r *URLRepository) GetByID(id string) (string, error) {
	for i, v := range r.store.urls {
		if id == v.URLShort {
			r.incrementStats(i, v)
			return v.URLOrigin, nil
		}
	}
	return "", fmt.Errorf("record not found")
}

func (r *URLRepository) incrementStats(i int, url model.URL) {
	r.store.Lock()
	defer r.store.Unlock()
	u := model.URL{
		URLShort:  url.URLShort,
		URLOrigin: url.URLOrigin,
		Visited:   true,
		Count:     url.Count + 1,
	}
	r.store.urls[i] = u
}
