package memstore_test

import (
	"errors"
	"github.com/iryzzh/practicum-go-shortener/internal/app/model"
	"github.com/iryzzh/practicum-go-shortener/internal/app/store"
	"github.com/iryzzh/practicum-go-shortener/internal/app/store/memstore"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUserRepository(t *testing.T) {
	st := memstore.New()

	url := model.TestURL(t)
	user := model.TestUser(t)

	assert.Condition(t, func() bool {
		err := st.URL().Create(url)
		if errors.Is(err, store.ErrURLExist) {
			return true
		}
		if err != nil {
			return false
		}

		return true
	})

	assert.NotNil(t, url.ID)

	assert.NoError(t, st.User().Create(user))
	assert.NotNil(t, user.ID)

	u, err := st.URL().FindByUUID(url.URLShort)
	assert.NoError(t, err)
	assert.Equal(t, url.URLOrigin, u.URLOrigin)

	u2, err := st.URL().FindByID(url.ID)
	assert.NoError(t, err)
	assert.Equal(t, u, u2)

	err = st.URL().UpdateUserID(url, user.ID)
	assert.NoError(t, err)
}
