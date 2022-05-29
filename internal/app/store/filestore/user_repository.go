package filestore

import (
	"encoding/json"
	"github.com/iryzzh/practicum-go-shortener/internal/app/model"
	"github.com/iryzzh/practicum-go-shortener/internal/app/store"
)

type UserRepository struct {
	store *Store
}

func (r *UserRepository) Create(user *model.User) error {
	if u, _ := r.FindByUUID(user.UUID); u != nil {
		return nil
	}

	r.store.Lock()
	user.ID = r.store.nextUserID + 1
	r.store.nextUserID++
	r.store.Unlock()

	data, err := json.Marshal(&user)
	if err != nil {
		return err
	}

	return r.store.Write(data, "user")
}

func (r *UserRepository) SaveURL(user *model.User, url *model.URL) error {
	if err := r.store.URL().Create(url); err != nil {
		return err
	}

	users, err := r.store.ReadUsers()
	if err != nil {
		return err
	}

	for _, v := range users {
		if v.UUID == user.UUID {
			for _, k := range v.URL {
				if k.ID == url.ID {
					return nil
				}
			}
			user.ID = v.ID
			user.URL = append(v.URL, model.UserURLID{ID: url.ID})
			b, err := json.Marshal(user)
			if err != nil {
				return err
			}
			return r.store.Write(b, "user")
		}
	}

	return store.ErrUserNotFound
}

func (r *UserRepository) FindByUUID(uuid string) (*model.User, error) {
	users, err := r.store.ReadUsers()
	if err != nil {
		return nil, err
	}

	for _, v := range users {
		if v.UUID == uuid {
			return &v, nil
		}
	}

	return nil, store.ErrUserNotFound
}

func (r *UserRepository) FindByID(id int) (*model.User, error) {
	users, err := r.store.ReadUsers()
	if err != nil {
		return nil, err
	}

	for _, v := range users {
		if v.ID == id {
			return &v, nil
		}
	}

	return nil, store.ErrUserNotFound
}
