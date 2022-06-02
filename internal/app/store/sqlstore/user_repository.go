package sqlstore

import (
	"database/sql"
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

	return r.store.db.QueryRow(
		"INSERT INTO users (uuid) VALUES ($1) RETURNING user_id",
		user.UUID,
	).Scan(&user.ID)
}

func (r *UserRepository) SaveURL(user *model.User, url *model.URL) error {
	if _, err := r.store.db.Exec(
		"UPDATE urls SET user_id = $1 WHERE url_id = $2",
		user.ID,
		url.ID,
	); err != nil {
		return err
	}

	return nil
}

func (r *UserRepository) FindByUUID(uuid string) (*model.User, error) {
	u := &model.User{}

	if err := r.store.db.QueryRow(
		"SELECT user_id, uuid FROM users WHERE uuid = $1",
		uuid,
	).Scan(
		&u.ID,
		&u.UUID,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, store.ErrRecordNotFound
		}

		return nil, err
	}

	return u, nil
}

func (r *UserRepository) FindByID(id int) (*model.User, error) {
	u := &model.User{}

	if err := r.store.db.QueryRow(
		"SELECT user_id, uuid FROM users WHERE user_id = $1",
		id,
	).Scan(
		&u.ID,
		&u.UUID,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, store.ErrRecordNotFound
		}

		return nil, err
	}

	return u, nil
}
