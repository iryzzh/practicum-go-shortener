package memstore_test

import (
	"github.com/iryzzh/practicum-go-shortener/internal/app/model"
	"github.com/iryzzh/practicum-go-shortener/internal/app/store/memstore"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUserCreate(t *testing.T) {
	store := memstore.New()
	user := model.TestUser(t)
	assert.NoError(t, store.User().Create(user))
	assert.NotNil(t, user.ID)
}

func TestUserSaveURL(t *testing.T) {
	store := memstore.New()
	user := model.TestUser(t)
	url := model.TestURL(t)
	assert.NoError(t, store.User().Create(user))
	assert.NoError(t, store.User().SaveURL(user, url))
}
