package memstore

import (
	"github.com/iryzzh/practicum-go-shortener/internal/app/model"
	"github.com/iryzzh/practicum-go-shortener/internal/app/store"
)

type UserRepository struct {
	store *Store
}

func (r *UserRepository) Create(user *model.User) error {
	r.store.Lock()
	defer r.store.Unlock()

	user.ID = r.store.userNextID + 1

	r.store.users[r.store.userNextID] = user
	r.store.userNextID++

	return nil
}

func (r *UserRepository) FindByID(id int) (*model.User, error) {
	for _, v := range r.store.users {
		if id == v.ID {
			return v, nil
		}
	}

	return nil, store.ErrUserNotFound
}

func (r *UserRepository) FindByUUID(uuid string) (*model.User, error) {
	for _, v := range r.store.users {
		if uuid == v.UUID {
			return v, nil
		}
	}

	return &model.User{}, store.ErrUserNotFound
}
