package filestore

import (
	"encoding/json"
	"github.com/iryzzh/practicum-go-shortener/internal/app/model"
	"github.com/iryzzh/practicum-go-shortener/internal/app/store"
	"log"
	"os"
	"sync"
)

type Store struct {
	sync.Mutex
	file   *os.File
	nextID int
}

func New(filepath string) (*Store, *os.File) {
	file, err := os.OpenFile(filepath, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0777)
	if err != nil {
		log.Fatal(err)
	}

	s := &Store{
		file:   file,
		nextID: 0,
	}

	s.startup()

	return s, file
}

func (s *Store) startup() {
	fi, err := s.file.Stat()
	if err != nil {
		log.Fatal(err)
	}
	r := NewReader(s.file, int(fi.Size()))
	data, err := r.LastLine()
	if err != nil {
		log.Fatal(err)
	}
	u := model.URL{}
	if len(data) > 0 {
		if err := json.Unmarshal([]byte(data), &u); err != nil {
			log.Fatal(err)
		}
		s.nextID = u.ID
	}
}

func (s *Store) URL() store.URLRepository {
	return &URLRepository{store: s}
}
