package memstore_test

import (
	"github.com/iryzzh/practicum-go-shortener/internal/app/model"
	"github.com/iryzzh/practicum-go-shortener/internal/app/store/memstore"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestURLRepository_Create(t *testing.T) {
	store := memstore.New()
	url := model.TestURL(t)
	assert.NoError(t, store.URL().Create(url))
	assert.NotNil(t, url.ID)
}

func TestURLRepository_FindByID(t *testing.T) {
	store := memstore.New()
	url := model.TestURL(t)
	assert.NoError(t, store.URL().Create(url))
	r, err := store.URL().FindByID(url.URLShort)
	assert.Equal(t, url.URLOrigin, r.URLOrigin)
	assert.NoError(t, err)
	assert.NotNil(t, r)
}
