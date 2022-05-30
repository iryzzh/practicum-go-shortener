package sqlstore

import "github.com/iryzzh/practicum-go-shortener/internal/app/model"

type URLRepository struct {
	store *Store
}

func (r *URLRepository) Exists(url *model.URL) bool {
	//TODO implement me
	return false
}

func (r *URLRepository) IncrementStats(i int) {
	//TODO implement me
}

func (r *URLRepository) Create(u *model.URL) error {
	return nil
}

func (r *URLRepository) FindByID(id int) (*model.URL, error) {
	return nil, nil
}

func (r *URLRepository) FindByUUID(uuid string) (*model.URL, error) {
	return nil, nil
}
