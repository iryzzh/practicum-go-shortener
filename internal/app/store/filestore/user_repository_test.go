package filestore_test

import (
	"github.com/iryzzh/practicum-go-shortener/internal/app/model"
	"github.com/iryzzh/practicum-go-shortener/internal/app/store/filestore"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUserRepository_Create(t *testing.T) {
	store, file := filestore.New(filepath)
	defer file.Close()
	user := model.TestUser(t)
	assert.NoError(t, store.User().Create(user))
	assert.NotNil(t, user.ID)
}

func TestUserRepository_SaveURL(t *testing.T) {
	store, file := filestore.New(filepath)
	defer file.Close()
	user := model.TestUser(t)
	url := model.TestURL(t)
	assert.NoError(t, store.User().Create(user))
	assert.NoError(t, store.User().SaveURL(user, url))
}
