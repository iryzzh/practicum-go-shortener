package memstore

import (
	"github.com/iryzzh/practicum-go-shortener/internal/app/model"
	"github.com/iryzzh/practicum-go-shortener/internal/app/store"
)

type URLRepository struct {
	store *Store
}

func (r *URLRepository) Create(url *model.URL) error {
	r.store.Lock()
	defer r.store.Unlock()
	if err := url.Validate(); err != nil {
		return err
	}

	for _, v := range r.store.urls {
		if url.URLOrigin == v.URLOrigin {
			*url = *v
			return store.ErrURLExist
		}
	}

	url.ID = r.store.urlNextID + 1

	r.store.urls[r.store.urlNextID] = url
	r.store.urlNextID++

	return nil
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

func (r *URLRepository) FindByUserID(id int) ([]*model.URL, error) {
	var result []*model.URL

	for _, v := range r.store.urls {
		if v.UserID == id {
			result = append(result, v)
		}
	}

	return result, nil
}

func (r *URLRepository) UpdateUserID(url *model.URL, userID int) error {
	url.UserID = userID

	return nil
}