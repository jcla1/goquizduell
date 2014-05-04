package util

import (
	"encoding/gob"
	"github.com/jcla1/goquizduell"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
)

func PrepareClient(username, password, cookieFileName string) (*quizduell.Client, error) {
	cookieURL, _ := url.Parse("https://qkgermany.feomedia.se/")
	var c *quizduell.Client
	var err error

	if cookie, err := loadCookie(cookieFileName); err == nil {
		jar, _ := cookiejar.New(nil)
		jar.SetCookies(cookieURL, []*http.Cookie{cookie})

		c, err = quizduell.NewClient(jar)
	} else {
		c, err = quizduell.NewClient(nil)
		if err != nil {
			return nil, err
		}

		_, err := c.Login(username, password)
		if err != nil {
			return nil, err
		}

		cookies := c.Jar.Cookies(cookieURL)
		if len(cookies) > 0 {
			// We'll just assume that the first
			// cookie is the auth cookie.
			cookie := cookies[0]

			err := saveCookie(cookieFileName, cookie)
			if err != nil {
				return nil, err
			}
		}
	}

	return c, err
}

func saveCookie(cookieFileName string, cookie *http.Cookie) error {
	file, err := os.Create(cookieFileName)
	if err != nil {
		return err
	}

	enc := gob.NewEncoder(file)
	err = enc.Encode(&cookie)
	if err != nil {
		return err
	}

	return nil
}

func loadCookie(cookieFileName string) (*http.Cookie, error) {
	file, err := os.Open(cookieFileName)
	if err != nil {
		return nil, err
	}

	dec := gob.NewDecoder(file)

	var cookie *http.Cookie
	err = dec.Decode(&cookie)
	if err != nil {
		return nil, err
	}

	return cookie, nil
}
