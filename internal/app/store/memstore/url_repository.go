package memstore

import (
	"github.com/iryzzh/practicum-go-shortener/internal/app/model"
	"github.com/iryzzh/practicum-go-shortener/internal/app/store"
)

type URLRepository struct {
	store *Store
}

func (r *URLRepository) IsDeleted(id int) bool {
	r.store.RLock()
	defer r.store.RUnlock()

	for _, v := range r.store.urls {
		if id == v.ID {
			return v.IsDeleted
		}
	}

	return false
}

func (r *URLRepository) BatchDelete(ids []int) error {
	r.store.Lock()
	defer r.store.Unlock()

	for i, v := range r.store.urls {
		for _, k := range ids {
			if k == v.ID {
				r.store.urls[i].IsDeleted = true
			}
		}
	}

	return nil
}

func (r *URLRepository) Delete(url *model.URL) error {
	r.store.Lock()
	defer r.store.Unlock()

	url.IsDeleted = true

	return nil
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
	r.store.RLock()
	defer r.store.RUnlock()

	for _, v := range r.store.urls {
		if id == v.ID {
			return v, nil
		}
	}

	return nil, store.ErrRecordNotFound
}

func (r *URLRepository) FindByUUID(uuid string) (*model.URL, error) {
	r.store.RLock()
	defer r.store.RUnlock()

	for _, v := range r.store.urls {
		if v.URLShort == uuid {
			return v, nil
		}
	}

	return nil, store.ErrRecordNotFound
}

func (r *URLRepository) FindByUserID(id int) ([]*model.URL, error) {
	r.store.RLock()
	defer r.store.RUnlock()

	var result []*model.URL

	for _, v := range r.store.urls {
		if v.UserID == id {
			result = append(result, v)
		}
	}

	return result, nil
}

func (r *URLRepository) UpdateUserID(url *model.URL, userID int) error {
	r.store.Lock()
	url.UserID = userID
	r.store.Unlock()

	return nil
}
