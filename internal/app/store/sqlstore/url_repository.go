package sqlstore

import (
	"database/sql"
	"github.com/iryzzh/practicum-go-shortener/internal/app/model"
	"github.com/iryzzh/practicum-go-shortener/internal/app/store"
)

type URLRepository struct {
	store *Store
}

func (r *URLRepository) Create(url *model.URL) error {
	if err := url.Validate(); err != nil {
		return err
	}

	shortURL := url.URLShort

	err := r.store.db.QueryRow(
		`WITH e AS (
    INSERT INTO urls ("original_url", "short_url")
        VALUES ($1, $2)
        ON CONFLICT ("original_url") DO NOTHING
        RETURNING "url_id", "short_url")
	SELECT *
	FROM e
	UNION
	SELECT "url_id", "short_url"
	FROM urls
	WHERE "original_url" = $1;`,
		url.URLOrigin,
		url.URLShort,
	).Scan(&url.ID, &url.URLShort)
	if err != nil {
		return err
	}

	if shortURL != url.URLShort {
		return store.ErrURLExist
	}

	return nil
}

func (r *URLRepository) FindByID(id int) (*model.URL, error) {
	u := &model.URL{}

	if err := r.store.db.QueryRow(
		"SELECT url_id, short_url, original_url FROM urls WHERE url_id = $1",
		id,
	).Scan(
		&u.ID,
		&u.URLShort,
		&u.URLOrigin,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, store.ErrRecordNotFound
		}

		return nil, err
	}

	return u, nil
}

func (r *URLRepository) FindByUUID(uuid string) (*model.URL, error) {
	u := &model.URL{}

	if err := r.store.db.QueryRow(
		"SELECT url_id, short_url, original_url FROM urls WHERE short_url = $1",
		uuid,
	).Scan(
		&u.ID,
		&u.URLShort,
		&u.URLOrigin,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, store.ErrRecordNotFound
		}

		return nil, err
	}

	return u, nil
}

func (r *URLRepository) FindByUserID(id int) ([]*model.URL, error) {
	var urls []*model.URL

	rows, err := r.store.db.Query(
		"SELECT * from urls where user_id = $1",
		id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var url model.URL
		if err := rows.Scan(&url.ID, &url.UserID, &url.URLOrigin, &url.URLShort); err != nil {
			return nil, err
		}

		urls = append(urls, &url)
	}

	return urls, nil
}

func (r *URLRepository) UpdateUserID(url *model.URL, userID int) error {
	url.UserID = userID

	if _, err := r.store.db.Exec(
		"UPDATE urls SET user_id = $1 WHERE url_id = $2",
		userID,
		url.ID); err != nil {
		return err
	}

	return nil
}
