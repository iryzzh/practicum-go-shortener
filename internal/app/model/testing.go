package model

import (
	"testing"
)

func TestURL(t *testing.T) *URL {
	t.Helper()

	return &URL{
		URLOrigin: "https://yandex.ru/pogoda/saint-petersburg",
		URLShort:  "gigsHiBV",
	}
}
