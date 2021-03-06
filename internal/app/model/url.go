package model

import (
	"fmt"
	"net/url"
	"regexp"
)

type URL struct {
	ID        int    `json:"id,omitempty"`
	URLOrigin string `json:"url"`
	URLShort  string `json:"url_short,omitempty"`
	UserID    int    `json:"user_id,omitempty"`
	IsDeleted bool   `json:"is_deleted"`
}

func (u *URL) Validate() error {
	reHTTP := regexp.MustCompile(`https?://`)
	if !reHTTP.MatchString(u.URLOrigin) {
		u.URLOrigin = "https://" + u.URLOrigin
	}
	t, err := url.Parse(u.URLOrigin)
	if err != nil {
		return err
	}

	reHost := regexp.MustCompile(`:.*`)
	reDomain := regexp.MustCompile(`^(?:[a-zA-Z\d](?:[a-zA-Z\d-]{0,61}[a-z\d])?\.)+(?:[a-zA-Z]{1,63}| xn--[a-z\d]{1,59})$`)
	host := reHost.ReplaceAllString(t.Host, "")

	if !reDomain.MatchString(host) {
		return fmt.Errorf("invalid host: %v", u.URLOrigin)
	}

	return nil
}
