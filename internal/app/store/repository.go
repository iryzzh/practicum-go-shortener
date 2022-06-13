package store

import "github.com/iryzzh/practicum-go-shortener/internal/app/model"

type URLRepository interface {
	Create(url *model.URL) error
	Delete(url *model.URL) error
	BatchDelete(ids []int) error
	FindByID(id int) (*model.URL, error)
	FindByUUID(uuid string) (*model.URL, error)
	FindByUserID(id int) ([]*model.URL, error)
	UpdateUserID(url *model.URL, userID int) error
}

type UserRepository interface {
	Create(user *model.User) error
	FindByUUID(uuid string) (*model.User, error)
	FindByID(id int) (*model.User, error)
}
