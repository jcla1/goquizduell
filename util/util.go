package util

import (
	"encoding/gob"
	"github.com/jcla1/goquizduell"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
)

func PrepareClient(username, password, cookieFileName string) *quizduell.Client {
	cookieURL, _ := url.Parse("https://qkgermany.feomedia.se/")
	var c *quizduell.Client

	if cookie, err := loadCookie(cookieFileName); err == nil {
		jar, _ := cookiejar.New(nil)
		jar.SetCookies(cookieURL, []*http.Cookie{cookie})

		c = quizduell.NewClient(jar)
	} else {
		c = quizduell.NewClient(nil)
		status := c.Login(username, password)

		if status == nil {
			return nil
		}

		cookies := c.Jar.Cookies(cookieURL)
		if len(cookies) > 0 {
			// We'll just assume that the first
			// cookie is the auth cookie.
			cookie := cookies[0]

			err := saveCookie(cookieFileName, cookie)
			if err != nil {
				panic(err)
			}
		}
	}

	return c
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
