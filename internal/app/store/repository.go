package store

import "github.com/iryzzh/practicum-go-shortener/internal/app/model"

type URLRepository interface {
	Create(url *model.URL) error
	FindById(string) (model.URL, error)
	GetById(string) (string, error)
}
