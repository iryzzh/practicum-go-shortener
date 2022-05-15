package filestore

import (
	"encoding/json"
	"errors"
	"github.com/iryzzh/practicum-go-shortener/internal/app/model"
	"io"
)

type URLRepository struct {
	store *Store
}

var (
	errRecordNotFound = errors.New("record not found")
)

func (r *URLRepository) Create(u *model.URL) error {
	if err := u.Validate(); err != nil {
		return err
	}

	file := r.store.file

	r.store.Lock()
	u.ID = r.store.nextID + 1
	r.store.nextID++
	r.store.Unlock()

	data, err := json.Marshal(&u)
	if err != nil {
		return err
	}

	data = append(data, '\n')

	_, err = file.Write(data)

	return err
}

func (r *URLRepository) FindByID(id string) (*model.URL, error) {
	m := &model.URL{}
	file := r.store.file
	fi, err := file.Stat()
	if err != nil {
		return m, err
	}

	rdr := NewReader(file, int(fi.Size()))
	for {
		line, err := rdr.LineReversed()
		if err != nil {
			if err == io.EOF {
				return m, errRecordNotFound
			}
			return m, err
		}
		if len(line) == 0 {
			continue
		}

		if err := json.Unmarshal([]byte(line), m); err != nil {
			return m, err
		}
		if id == m.URLShort {
			return m, nil
		}
	}
}

func (r *URLRepository) IncrementStats(i int) {}
