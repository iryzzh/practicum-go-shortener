package memstore_test

import (
	"github.com/iryzzh/practicum-go-shortener/internal/app/model"
	"github.com/iryzzh/practicum-go-shortener/internal/app/store/memstore"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestURLRepository(t *testing.T) {
	store := memstore.New()

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
