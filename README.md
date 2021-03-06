Quizduell API client
===========

Inofficial API client for the Quizduell REST API written in Go.
Inspired by [https://github.com/mtschirs/quizduellapi](https://github.com/mtschirs/quizduellapi) and [easysurfer.me](http://easysurfer.me/wordpress/?p=761)

For documentation, [see here](http://godoc.org/github.com/jcla1/goquizduell).

## Use cases
This API library supports all functionality that can be found in the official Quizduell app, besides posting new questions. There is also no need to have downloaded, or registered a premium account in order to use features in this client library that would otherwise be unavailable to a normal user (like setting a new avatar and playing more than 8 games simulanously).

Included in the repo is a demo of the API in which an automated player logs into Quizduell and plays active games, accepts game invites and keeps a constant pool of games against random opponents. It also supports things like: adjusting the standard deviation from the correct answer (how often does the player answer correctly), excluding players on your friends list from gameplay against the automated player and a few other things.

To automate the process of starting the player, there is [a script](runner.sh) included that starts the player every few minutes with a cookie in `cookie.gob` (which you will have to create yourself). To run it, just do: `./runner.sh`

When starting the player for the first time, you'll have to provide your username, password and a name for the cookie file in environmental variables named `QD_USERNAME`, `QD_PASSWORD`, `QD_COOKIE_FILE`, respectively. After this initial run, you'll only ever have to provide the `QD_COOKIE_FILE` variable, provided the file actually still exists.

__*Warning*__: Most calls to the API do __not__ populate all fields of the returned structs. Some may still contain their respective null values. i.e. when retrieving the list of active games from a `*quizduell.Status` the list of questions for a particular game will be empty.

__*Note*__: When calling `func (c *Client) UploadRoundAnswers` you must also pass all of your previous answers to the questions, in addition to your answers of this round. If you are passing the answers to questions where your opponent choose the category, you must also pass the correct category for those questions/answers (i.e. `game.CategoryChoices[len(game.CategoryChoices)-1]`).

## License
Pubished under the [MIT License](LICENSE).

Quizduell is a registered trademark of FEO Media AB, Stockholm, SE registered in Germany and other countries. This project is an independent work and is in no way affiliated with, authorized, maintained, sponsored or endorsed by FEO Media AB.
