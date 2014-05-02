package main

import (
	"flag"
	"fmt"
	"github.com/jcla1/goquizduell"
	"github.com/jcla1/goquizduell/util"
	"math"
	"math/rand"
	"os"
	"strings"
)

var numRandGames = flag.Int("rand-games", 0, "number of random games to start")
var constGames = flag.Int("const-games", 20, "how many random games to maintain")
var ansStdDev = flag.Float64("ans-stddev", 0.8, "parameter to control the number of correct answers the player gives")
var giveUpMins = flag.Int("give-up-mins", 360, "number of minutes to play a game before giving up")

var noPlayNames stringSlice

func init() {
	flag.Var(&noPlayNames, "no-play-names", "comma-seperated list of usernames that should not be played against")
}

func main() {
	flag.Parse()
	c := util.PrepareClient(os.Getenv("QD_USERNAME"), os.Getenv("QD_PASSWORD"), os.Getenv("QD_COOKIE_FILE"))

	games := c.GetUserGames().User.Games

	activeGameCount := 0

	for _, game := range games {
		if isNoPlayName(game.Opponent.Name) {
			continue
		}

		if game.GameState == quizduell.Active {
			activeGameCount += 1

			if !game.YourTurn && game.ElapsedMinutes > *giveUpMins {
				activeGameCount -= 1
				fmt.Println("Giving up game against:", game.Opponent.Name)
				c.GiveUp(game.ID)
			}
		}

		// First we accept any game requests
		if game.GameState == quizduell.Waiting && game.YourTurn {
			activeGameCount += 1
			fmt.Println("Accepting invite from:", game.Opponent.Name)
			c.AcceptGame(game.ID)
		}

		// Answer the questions
		if game.YourTurn {
			numAns := findNumRequiredAns(game)
			categoryID := findCorrectCategoryID(game, numAns)
			answers := make([]int, numAns)

			if numAns == 3 && len(game.OpponentAnswers) > 0 {
				activeGameCount -= 1
			}

			correctCount := 0
			for i := range answers {
				answers[i] = randAnswer()
				if answers[i] == 0 {
					correctCount++
				} else if answers[i] != 1 && answers[i] != 2 && answers[i] != 3 {
					panic(answers[i])
				}
			}

			fmt.Println("Answering", numAns, "questions against:", game.Opponent.Name, "[ correct:", correctCount, "]")
			c.UploadRoundAnswers(game.ID, append(game.YourAnswers, answers...), categoryID)
		}
	}

	stats := c.CategoryStatistics()
	fmt.Println("---\nCurrently playing in", activeGameCount, "active games.")
	fmt.Printf("My current rank is %d/%d users.\n", stats.Rank, stats.NumUsers)
	// fmt.Println("So far you I have won", stats.GamesWon, "games with", stats.QuestionsCorrect, "correct answers.")
	fmt.Println("---")

	gamesToStart := *numRandGames

	if (activeGameCount + gamesToStart) < *constGames {
		gamesToStart += *constGames - (activeGameCount + gamesToStart)
	}

	for i := 0; i < gamesToStart; i++ {
		g := c.StartRandomGame()
		fmt.Println("Starting random game against:", g.Opponent.Name)
	}
}

func randAnswer() int {
	return int(math.Abs(rand.NormFloat64() * (*ansStdDev)))
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

func isNoPlayName(name string) bool {
	for _, other := range noPlayNames {
		if name == other {
			return true
		}
	}
	return false
}

type stringSlice []string

func (s *stringSlice) String() string {
	return fmt.Sprint(*s)
}

func (s *stringSlice) Set(value string) error {
	for _, name := range strings.Split(value, ",") {
		name = strings.Trim(name, " ")
		*s = append(*s, name)
	}
	return nil
}
