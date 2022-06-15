package sqlstore

import (
	"database/sql"
	"github.com/iryzzh/practicum-go-shortener/internal/app/model"
	"github.com/iryzzh/practicum-go-shortener/internal/app/store"
	"github.com/lib/pq"
	"github.com/pkg/errors"
)

type URLRepository struct {
	store *Store
}

func (r *URLRepository) BatchDelete(ids []int) error {
	_, err := r.store.db.Exec(`UPDATE urls SET is_deleted = true WHERE url_id = ANY($1::int[]);`, pq.Array(ids))

	return err
}

func (r *URLRepository) Delete(url *model.URL) error {
	_, err := r.store.db.Exec(`UPDATE urls SET is_deleted = true WHERE url_id = $1`, url.ID)

	return err
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

	err := r.store.db.QueryRow(
		"SELECT url_id, short_url, original_url FROM urls WHERE url_id = $1",
		id,
	).Scan(
		&u.ID,
		&u.URLShort,
		&u.URLOrigin,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, store.ErrRecordNotFound
	}
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (r *URLRepository) FindByUUID(uuid string) (*model.URL, error) {
	u := &model.URL{}

	err := r.store.db.QueryRow(
		"SELECT url_id, user_id, original_url, short_url, is_deleted FROM urls WHERE short_url = $1",
		uuid,
	).Scan(
		&u.ID,
		&u.UserID,
		&u.URLOrigin,
		&u.URLShort,
		&u.IsDeleted,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, store.ErrRecordNotFound
	}
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (r *URLRepository) FindByUserID(id int) ([]*model.URL, error) {
	var urls []*model.URL

	rows, err := r.store.db.Query(
		"SELECT url_id, user_id, original_url, short_url, is_deleted from urls where user_id = $1",
		id)
	if err != nil {
		return nil, errors.Wrap(err, "query")
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var url model.URL
		if err := rows.Scan(&url.ID, &url.UserID, &url.URLOrigin, &url.URLShort, &url.IsDeleted); err != nil {
			return nil, errors.Wrap(err, "scan rows")
		}

		urls = append(urls, &url)
	}

	return urls, rows.Err()
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
