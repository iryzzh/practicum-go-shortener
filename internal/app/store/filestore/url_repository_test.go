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

func TestURLRepository(t *testing.T) {
	store, file := filestore.New(filepath)
	defer file.Close()

	url := model.TestURL(t)
	user := model.TestUser(t)

	assert.NoError(t, store.URL().Create(url))
	assert.NotNil(t, url.ID)

	assert.NoError(t, store.User().Create(user))
	assert.NotNil(t, user.ID)

	u, err := store.URL().FindByUUID(url.URLShort)
	assert.NoError(t, err)
	assert.Equal(t, url.URLOrigin, u.URLOrigin)

	u2, err := store.URL().FindByID(url.ID)
	assert.NoError(t, err)
	assert.Equal(t, u, u2)

	err = store.URL().UpdateUserID(url, user.ID)
	assert.NoError(t, err)

	urls, err := store.URL().FindByUserID(user.ID)
	assert.NoError(t, err)
	assert.NotNil(t, urls)
}
