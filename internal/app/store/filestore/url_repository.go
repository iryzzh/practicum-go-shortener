package filestore

import (
	"github.com/iryzzh/practicum-go-shortener/internal/app/model"
	"github.com/iryzzh/practicum-go-shortener/internal/app/store"
)

type URLRepository struct {
	store *Store
}

func (r *URLRepository) BatchDelete(ids []int) error {
	urls, err := r.store.ReadUrls()
	if err != nil {
		return err
	}

	for _, v := range urls {
		for _, k := range ids {
			if k == v.ID {
				v.IsDeleted = true

				b, err := json.Marshal(v)
				if err != nil {
					return err
				}

				if err := r.store.Write(b, "url"); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (r *URLRepository) Delete(url *model.URL) error {
	urls, err := r.store.ReadUrls()
	if err != nil {
		return err
	}

	for _, v := range urls {
		if url.URLOrigin == v.URLOrigin {
			if v.IsDeleted {
				return nil
			}

			v.IsDeleted = true

			b, err := json.Marshal(v)
			if err != nil {
				return err
			}

			return r.store.Write(b, "url")
		}
	}

	return store.ErrRecordNotFound
}

func (r *URLRepository) Create(url *model.URL) error {
	if err := url.Validate(); err != nil {
		return err
	}

	urls, err := r.store.ReadUrls()
	if err != nil {
		return err
	}

	for _, v := range urls {
		if url.URLOrigin == v.URLOrigin {
			*url = v
			return store.ErrURLExist
		}
	}

	r.store.Lock()
	url.ID = r.store.nextURLID + 1
	r.store.nextURLID++
	r.store.Unlock()

	data, err := json.Marshal(&url)
	if err != nil {
		return err
	}

	return r.store.Write(data)
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

func (r *URLRepository) FindByUserID(id int) ([]*model.URL, error) {
	var result []*model.URL

	urls, err := r.store.ReadUrls()
	if err != nil {
		return nil, err
	}

	for _, v := range urls {
		if v.UserID == id {
			result = append(result, &v)
		}
	}

	return result, nil
}

func (r *URLRepository) UpdateUserID(url *model.URL, userID int) error {
	url.UserID = userID

	u, err := r.store.URL().FindByUserID(userID)
	if err != nil {
		return err
	}

	for _, v := range u {
		if v.URLShort == url.URLShort && v.UserID == url.UserID {
			return nil
		}
	}

	b, err := json.Marshal(url)
	if err != nil {
		return err
	}

	return r.store.Write(b, "url")
}
