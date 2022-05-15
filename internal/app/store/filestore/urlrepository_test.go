package filestore_test

import (
	"github.com/iryzzh/practicum-go-shortener/internal/app/model"
	"github.com/iryzzh/practicum-go-shortener/internal/app/store/filestore"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"testing"
)

var (
	file = "testfile.txt"
)

func TestURLRepository_Create(t *testing.T) {
	file, err := os.OpenFile(file, os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	store := filestore.New(file)
	url := model.TestURL(t)
	assert.NoError(t, store.URL().Create(url))
	assert.NotNil(t, url.ID)
}

func TestURLRepository_FindByID(t *testing.T) {
	file, err := os.OpenFile(file, os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	store := filestore.New(file)
	url := model.TestURL(t)
	assert.NoError(t, store.URL().Create(url))
	r, err := store.URL().FindByID(url.URLShort)
	assert.Equal(t, url.URLOrigin, r.URLOrigin)
	assert.NoError(t, err)
	assert.NotNil(t, r)
}
