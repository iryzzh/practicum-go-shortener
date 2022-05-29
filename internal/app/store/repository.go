package store

import "github.com/iryzzh/practicum-go-shortener/internal/app/model"

type URLRepository interface {
	Create(url *model.URL) error
	FindByID(id int) (*model.URL, error)
	FindByUUID(uuid string) (*model.URL, error)
	Exists(url *model.URL) bool
	IncrementStats(int)
}

type UserRepository interface {
	Create(user *model.User) error
	SaveURL(user *model.User, url *model.URL) error
	FindByUUID(uuid string) (*model.User, error)
	FindByID(id int) (*model.User, error)
}
