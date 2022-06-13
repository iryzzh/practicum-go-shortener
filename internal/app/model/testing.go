package model

import (
	"fmt"
	"github.com/iryzzh/practicum-go-shortener/internal/pkg/utils"
	"math/rand"
	"testing"
	"time"
)

func TestURL(t *testing.T) *URL {
	t.Helper()

	return &URL{
		URLOrigin: "https://yandex.ru/pogoda/saint-petersburg",
		URLShort:  "g1gsHibv",
	}
}

func TestURLGenerated(t *testing.T) *URL {
	t.Helper()

	rand.Seed(time.Now().UnixNano())

	min := 5
	max := 10

	var result []string
	for i := 0; i < 3; i++ {
		result = append(result, utils.RandStringLowerCase(rand.Intn(max-min+1)+min))
	}

	urlStr := fmt.Sprintf("http://%s.ru/%s/%s", result[0], result[1], result[2])

	url := &URL{
		URLOrigin: urlStr,
		URLShort:  utils.RandString(8),
	}

	return url
}

func TestUser(t *testing.T) *User {
	t.Helper()

	return &User{
		UUID: "353ba025-7285-4790-bfeb-b70c1ef18323",
	}
}
