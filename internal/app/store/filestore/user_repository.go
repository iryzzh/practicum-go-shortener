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
		user.ID = u.ID
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
