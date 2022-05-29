package filestore

import (
	"encoding/json"
	"github.com/iryzzh/practicum-go-shortener/internal/app/model"
	"github.com/iryzzh/practicum-go-shortener/internal/app/store"
)

type URLRepository struct {
	store *Store
}

func (r *URLRepository) Create(u *model.URL) error {
	if err := u.Validate(); err != nil {
		return err
	}

	if _, err := r.store.URL().FindByUUID(u.URLShort); err != nil && err != store.ErrRecordNotFound {
		return err
	}

	r.store.Lock()
	u.ID = r.store.nextUrlID + 1
	r.store.nextUrlID++
	r.store.Unlock()

	data, err := json.Marshal(&u)
	if err != nil {
		return err
	}

	return r.store.Write(data)
}

func (r *URLRepository) Exists(u *model.URL) bool {
	urls, err := r.store.ReadUrls()
	if err != nil {
		return false
	}

	for _, v := range urls {
		if v.URLShort == u.URLShort {
			return true
		}
	}

	return false
}

func (r *URLRepository) FindByID(id int) (*model.URL, error) {
	urls, err := r.store.ReadUrls()
	if err != nil {
		return nil, err
	}

	for _, v := range urls {
		if v.ID == id {
			return &v, nil
		}
	}

	return nil, store.ErrRecordNotFound
}

func (r *URLRepository) FindByUUID(uuid string) (*model.URL, error) {
	urls, err := r.store.ReadUrls()
	if err != nil {
		return nil, err
	}

	for _, v := range urls {
		if v.URLShort == uuid {
			return &v, nil
		}
	}

	return nil, store.ErrRecordNotFound
}

func (r *URLRepository) IncrementStats(i int) {}
