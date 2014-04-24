package main

import (
	"encoding/gob"
	"fmt"
	"github.com/jcla1/goquizduell"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
)

func main() {
	c := prepareClient(os.Getenv("QD_USERNAME"), os.Getenv("QD_PASSWORD"), os.Getenv("QD_COOKIE_FILE"))

	games := c.GetUserGames().User.Games

	for _, game := range games {
		// First we accept any game requests
		if game.GameState == quizduell.Waiting && game.YourTurn {
			fmt.Println("Accepting invite from: ", game.Opponent.Name)
			c.AcceptGame(game.ID)
		}

		// Answer the questions
		if game.YourTurn {
			numAns := findNumRequiredAns(game)
			categoryID := findCorrectCategoryID(game, numAns)
			answers := make([]int, numAns)

			fmt.Println("Answering", numAns, "questions against:", game.Opponent.Name)
			c.UploadRoundAnswers(game.ID, append(game.YourAnswers, answers...), categoryID)
		}
	}
}

func findCorrectCategoryID(game quizduell.Game, numAns int) int {
	if numAns == 3 && len(game.OpponentAnswers) != 0 {
		return game.CategoryChoices[len(game.CategoryChoices)-1]
	}
	// We don't care what category we choose otherwise!
	return 0
}

func findNumRequiredAns(game quizduell.Game) int {
	if len(game.OpponentAnswers) == 0 || len(game.OpponentAnswers) == 18 {
		return 3
	}
	return 6
}

func prepareClient(username, password, cookieFileName string) *quizduell.Client {
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
