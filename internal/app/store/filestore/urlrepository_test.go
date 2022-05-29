package filestore_test

import (
	"github.com/iryzzh/practicum-go-shortener/internal/app/model"
	"github.com/iryzzh/practicum-go-shortener/internal/app/store/filestore"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	filepath = "testfile.txt"
)

func TestURLRepository_Create(t *testing.T) {
	store, file := filestore.New(filepath)
	defer file.Close()
	url := model.TestURL(t)
	assert.NoError(t, store.URL().Create(url))
	assert.NotNil(t, url.ID)
}

func TestURLRepository_FindByID(t *testing.T) {
	store, file := filestore.New(filepath)
	defer file.Close()
	url := model.TestURL(t)
	assert.NoError(t, store.URL().Create(url))
	r, err := store.URL().FindByUUID(url.URLShort)
	assert.Equal(t, url.URLOrigin, r.URLOrigin)
	assert.NoError(t, err)
	assert.NotNil(t, r)
}
