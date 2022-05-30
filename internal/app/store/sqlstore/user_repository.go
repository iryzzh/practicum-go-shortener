package sqlstore

import (
	"github.com/iryzzh/practicum-go-shortener/internal/app/model"
)

type UserRepository struct {
	store *Store
}

func (u UserRepository) Create(user *model.User) error {
	//TODO implement me
	return nil
}

func (u UserRepository) SaveURL(user *model.User, url *model.URL) error {
	//TODO implement me
	return nil
}

func (u UserRepository) FindByUUID(uuid string) (*model.User, error) {
	//TODO implement me
	return nil, nil
}

func (u UserRepository) FindByID(id int) (*model.User, error) {
	//TODO implement me
	return nil, nil
}
