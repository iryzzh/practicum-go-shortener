package model

import (
	"testing"
)

func TestURL(t *testing.T) *URL {
	t.Helper()

	return &URL{
		URLOrigin: "https://yandex.ru/pogoda/saint-petersburg",
		URLShort:  "g1gsHibv",
	}
}

func TestUser(t *testing.T) *User {
	t.Helper()

	return &User{
		UUID: "353ba025-7285-4790-bfeb-b70c1ef18323",
	}
}
