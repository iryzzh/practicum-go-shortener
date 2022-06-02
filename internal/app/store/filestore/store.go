package filestore

import (
	"github.com/iryzzh/practicum-go-shortener/internal/app/model"
	"github.com/iryzzh/practicum-go-shortener/internal/app/store"
	"github.com/json-iterator/go"
	"io"
	"log"
	"os"
	"sync"
)

var (
	json = jsoniter.ConfigCompatibleWithStandardLibrary
)

type Store struct {
	sync.Mutex
	fileDescriptor *os.File
	nextUrlID      int
	nextUserID     int
	fileData       File
}

func (s *Store) Close() error {
	return s.fileDescriptor.Close()
}

func New(filepath string) (*Store, error) {
	file, err := os.OpenFile(filepath, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0777)
	if err != nil {
		return nil, err
	}

	s := &Store{
		fileDescriptor: file,
		nextUrlID:      0,
		nextUserID:     0,
	}

	s.startup()

	return s, nil
}

func (s *Store) startup() {
	data, err := s.Read()
	if err != nil {
		log.Fatal(err)
	}

	var user model.User
	var url model.URL
	var f File

	for _, v := range data {
		if s.nextUserID > 0 && s.nextUrlID > 0 {
			break
		}
		if err := json.Unmarshal([]byte(v), &f); err == nil {
			if s.nextUserID == 0 {
				if f.Type == "user" {
					if err := json.Unmarshal(f.Data, &user); err == nil {
						s.nextUserID = user.ID
					}
				}
			}
			if s.nextUrlID == 0 {
				if f.Type == "url" || f.Type == "" {
					if err := json.Unmarshal(f.Data, &url); err == nil {
						s.nextUrlID = url.ID
					}
				}
			}
		}
	}
}

type File struct {
	Type string              `json:"type"`
	Data jsoniter.RawMessage `json:"data"`
}

func (s *Store) Write(data []byte, dataType ...string) error {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	dt := "url"
	if len(dataType) > 0 {
		dt = dataType[0]
	}

	f := &File{
		Data: jsoniter.RawMessage(data),
		Type: dt,
	}

	b, err := json.Marshal(f)
	if err != nil {
		return err
	}

	b = append(b, '\n')

	_, err = s.fileDescriptor.Write(b)

	return err
}

func (s *Store) ReadUsers() ([]model.User, error) {
	f := File{}
	u := model.User{}
	var users []model.User

	data, err := s.Read()
	if err != nil {
		return nil, err
	}

	for _, v := range data {
		if err := json.Unmarshal([]byte(v), &f); err == nil {
			if f.Type == "user" {
				if err := json.Unmarshal(f.Data, &u); err == nil {
					users = append(users, u)
				}
			}
		}
	}

	return users, nil
}

func (s *Store) ReadUrls() ([]model.URL, error) {
	f := File{}
	u := model.URL{}
	var urls []model.URL

	data, err := s.Read()
	if err != nil {
		return nil, err
	}

	for _, v := range data {
		if err := json.Unmarshal([]byte(v), &f); err == nil {
			if f.Type != "user" {
				if err := json.Unmarshal(f.Data, &u); err == nil {
					urls = append(urls, u)
				}
			}
		}
	}

	return urls, nil
}

func (s *Store) Read() ([]string, error) {
	reader, err := s.Reader()
	if err != nil {
		return nil, err
	}

	var lines []string

	for {
		line, err := reader.LineReversed()
		if err != nil {
			if err == io.EOF {
				return lines, nil
			}
			return nil, err
		}
		if len(line) == 0 {
			continue
		}

		lines = append(lines, line)
	}
}

func (s *Store) Reader() (*Reader, error) {
	file := s.fileDescriptor
	fi, err := file.Stat()
	if err != nil {
		return nil, err
	}

	return NewReader(file, int(fi.Size())), nil
}

func (s *Store) URL() store.URLRepository {
	return &URLRepository{store: s}
}

func (s *Store) User() store.UserRepository {
	return &UserRepository{store: s}
}

func (s *Store) Ping() error {
	return nil
}
