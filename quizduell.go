// Package quizduell provides an interface to the
// REST API used by the Quizduell mobile apps.
// It supports all functionality that is also
// possible in the mobile apps.
// Note: Most calls to the API do _not_ populate all
//       fields of the returned type.
//       As an example, Client.GetUserGames() does not
//       include the full-text of the questions.
package quizduell

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	protocolPrefix = "https://"
	hostName       = "qkgermany.feomedia.se"
	userAgent      = "Quizduell A 1.3.2"
	authKey        = "irETGpoJjG57rrSC"
	passwordSalt   = "SQ2zgOTmQc8KXmBP"
)

// Client represents a single user's (non-persistent)
// connection to Quizduell. In fact it just wraps
// the HTTP client and cookiejar.
type Client struct {
	client *http.Client
	// We're separating out this cookie jar from
	// the HTTP client, because Go is trying to be
	// peticularly RFC compliant and doesn't allow
	// certain characters in cookies. That's why we
	// have to handle them seperately.
	Jar http.CookieJar
}

// NewClient produces a new Quizduell client. It
// optionally takes a cookiejar, but if there isn't
// one provided it automatically creates one for you.
func NewClient(cookieJar http.CookieJar) (*Client, error) {
	if cookieJar == nil {
		jar, err := cookiejar.New(nil)
		if err != nil {
			return nil, err
		}

		cookieJar = jar
	}

	return &Client{
		Jar: cookieJar,
		client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		},
	}, nil
}

// Login logs in a user to Quizduell and puts the
// returned cookie (on success) into the cookiejar.
// You've no need to call login, if you create a
// new user or provide a new cookiejar with the
// appropriate cookie in it.
func (c *Client) Login(username, password string) (*Status, error) {
	data := url.Values{}

	h := md5.New()
	io.WriteString(h, passwordSalt+password)
	data.Set("pwd", string(hex.EncodeToString(h.Sum(nil))))
	data.Set("name", username)

	msg, err := c.makeRequest("/users/login", data)
	if err == nil {
		return msg.Status, nil
	}

	return nil, err
}

// CreateUser registers a new user with Quizduell,
// this user is automatically logged in. The email
// is optional and will be omitted for the call if
// it is the empty string.
func (c *Client) CreateUser(username, email, password string) (*Status, error) {
	data := url.Values{}

	data.Set("name", username)
	if email != "" {
		data.Set("email", email)
	}

	h := md5.New()
	io.WriteString(h, passwordSalt+password)
	data.Set("pwd", string(hex.EncodeToString(h.Sum(nil))))

	msg, err := c.makeRequest("/users/create", data)
	if err == nil {
		return msg.Status, nil
	}

	return nil, err
}

// UpdateUser sets the user's attributes, if one of
// them is the empty string that attribute will be
// omitted from the request.
// Requires you to be logged in.
func (c *Client) UpdateUser(username, email, password string) (*Status, error) {
	data := url.Values{}

	if username != "" {
		data.Set("name", username)
	}

	if email != "" {
		data.Set("email", email)
	}

	if password != "" {
		h := md5.New()
		io.WriteString(h, passwordSalt+password)
		data.Set("pwd", string(hex.EncodeToString(h.Sum(nil))))
	}

	msg, err := c.makeRequest("/users/update_user", data)
	if err == nil {
		return msg.Status, nil
	}

	return nil, err
}

// CreateTVUser creates a new TV user profile for the
// current logged in Quizduell users.
// Requires you to be logged in.
func (c *Client) CreateTVUser() (*User, error) {
	msg, err := c.makeRequest("/tv/create_tv_user", url.Values{})
	if err == nil {
		return msg.User, nil
	}

	return nil, err
}

// FindUser returns the user object of the user
// with the provided username.
// Requires you to be logged in.
func (c *Client) FindUser(username string) (*User, error) {
	data := url.Values{}
	data.Set("opponent_name", username)

	msg, err := c.makeRequest("/users/find_user", data)
	if err == nil {
		return msg.U, nil
	}

	return nil, err
}

// AddFriend puts the user with the provided userID onto
// your friends list.
// Requires you to be logged in.
func (c *Client) AddFriend(userID int) (*Popup, error) {
	data := url.Values{}

	data.Set("friend_id", strconv.Itoa(userID))

	msg, err := c.makeRequest("/users/add_friend", data)
	if err == nil {
		return msg.Popup, nil
	}

	return nil, err
}

// RemoveFriend removes the user with the provided userID
// from your friends list.
// Requires you to be logged in.
func (c *Client) RemoveFriend(userID int) (*Popup, error) {
	data := url.Values{}

	data.Set("friend_id", strconv.Itoa(userID))

	msg, err := c.makeRequest("/users/remove_friend", data)
	if err == nil {
		return msg.Popup, nil
	}

	return nil, err
}

// UpdateAvatar sets the current user's avatar to the provided
// avatar code. An avatar consists of individual mouth, hair,
// eyes, hats, etc. encoded in a numerical string, e.g. "0010999912"
// (A skin-colored avatar with a crown).
// Requires you to be logged in.
func (c *Client) UpdateAvatar(avatarCode string) (bool, error) {
	data := url.Values{}

	data.Set("avatar_code", avatarCode)

	msg, err := c.makeRequest("/users/update_avatar", data)
	if err == nil {
		return msg.T, nil
	}

	return false, err
}

// SendForgotPasswordEmail sends a forgot password email to the
// current user. I guess this function requires the current user
// to have an email set in his profile.
// Requires you to be logged in.
func (c *Client) SendForgotPasswordEmail(email string) (*Popup, error) {
	data := url.Values{}

	data.Set("email", email)

	msg, err := c.makeRequest("/users/forgot_pwd", data)
	if err == nil {
		return msg.Popup, nil
	}

	return nil, err
}

// AddBlocked puts the user with the provided userID onto
// your blocked list.
// Requires you to be logged in.
func (c *Client) AddBlocked(userID int) ([]User, error) {
	data := url.Values{}

	data.Set("blocked_id", strconv.Itoa(userID))

	msg, err := c.makeRequest("/users/add_blocked", data)
	if err == nil {
		return msg.Blocked, nil
	}

	return nil, err
}

// RemoveBlocked removes the user with the provided userID
// from your blocked list.
// Requires you to be logged in.
func (c *Client) RemoveBlocked(userID int) ([]User, error) {
	data := url.Values{}

	data.Set("blocked_id", strconv.Itoa(userID))

	msg, err := c.makeRequest("/users/remove_blocked", data)
	if err == nil {
		return msg.Blocked, nil
	}

	return nil, err
}

// StartGame starts a new game against the player with
// the provided opponentID.
// Requires you to be logged in.
func (c *Client) StartGame(opponentID int) (*Game, error) {
	data := url.Values{}

	data.Set("opponent_id", strconv.Itoa(opponentID))

	msg, err := c.makeRequest("/games/create_game", data)
	if err == nil {
		return msg.Game, nil
	}

	return nil, err
}

// StartRandomGame starts a new game against a player
// that is choosen randomly by the Quizduell server.
// After some time testing the automatic player, it
// seems that you can at most play in about 122 games.
// Requires you to be logged in.
func (c *Client) StartRandomGame() (*Game, error) {
	msg, err := c.makeRequest("/games/start_random_game", nil)
	if err == nil {
		return msg.Game, nil
	}

	return nil, err
}

// GetGame returns more information about the game with
// the given gameID. The returned game object also contains
// the all possible questions of every round.
// Requires you to be logged in.
func (c *Client) GetGame(gameID int) (*Game, error) {
	msg, err := c.makeRequest("/games/"+strconv.Itoa(gameID), nil)
	if err == nil {
		return msg.Game, nil
	}

	return nil, err
}

// GetGames returns details of the specified games, not
// including question and answer strings though.
// Requires you to be logged in.
func (c *Client) GetGames(gameIDs []int) ([]Game, error) {
	data := url.Values{}

	data.Set("gids", stringifyIntSlice(gameIDs))

	msg, err := c.makeRequest("/games/short_games", data)
	if err == nil {
		return msg.Games, nil
	}

	return nil, err
}

// GiveUp ends the game with the provided gameID, you may
// loose points when giving up.
// Requires you to be logged in.
func (c *Client) GiveUp(gameID int) (*Game, *Popup, error) {
	data := url.Values{}

	data.Set("game_id", strconv.Itoa(gameID))

	msg, err := c.makeRequest("/games/give_up", data)
	if err == nil {
		return msg.Game, msg.Popup, nil
	}

	return nil, nil, err
}

// AcceptGame accepts a pending game request that has the
// given gameID.
// Requires you to be logged in.
func (c *Client) AcceptGame(gameID int) (bool, error) {
	data := url.Values{}

	data.Set("accept", "1")
	data.Set("game_id", strconv.Itoa(gameID))

	msg, err := c.makeRequest("/games/accept", data)
	if err == nil {
		return msg.T, nil
	}

	return false, err
}

// DeclineGame declines a pending game request that has the
// given gameID.
// Requires you to be logged in.
func (c *Client) DeclineGame(gameID int) (bool, error) {
	data := url.Values{}

	data.Set("accept", "0")
	data.Set("game_id", strconv.Itoa(gameID))

	msg, err := c.makeRequest("/games/accept", data)
	if err == nil {
		return msg.T, nil
	}

	return false, err
}

// UploadRoundAnswers sends your provided answers to the
// Quizduell server.
// Note: In the answers you must include all answers you
// gave in the previous rounds of the same game.
// Requires you to be logged in.
func (c *Client) UploadRoundAnswers(gameID int, answers []int, categoryID int) (*Game, error) {
	data := url.Values{}

	data.Set("game_id", strconv.Itoa(gameID))
	data.Set("cat_choice", strconv.Itoa(categoryID))
	data.Set("answers", stringifyIntSlice(answers))

	msg, err := c.makeRequest("/games/upload_round_answers", data)
	if err == nil {
		return msg.Game, nil
	}

	return nil, err
}

// GetUserGames returns a status update, that also contains
// game data from the user's games.
// Requires you to be logged in.
func (c *Client) GetUserGames() (*Status, error) {
	msg, err := c.makeRequest("/users/current_user_games", url.Values{})
	if err == nil {
		return msg.Status, nil
	}

	return nil, err
}

// SendMessage sends a message to the user that is the opponent
// in the game with the given gameID. All messages to a user
// are visible in all games against this opponent.
// Requires you to be logged in.
func (c *Client) SendMessage(gameID int, message string) (*InGameMessage, error) {
	data := url.Values{}

	data.Set("game_id", strconv.Itoa(gameID))
	data.Set("text", message)

	msg, err := c.makeRequest("/games/send_message", nil)
	if err == nil {
		return msg.InGameMessage, nil
	}

	return nil, err
}

// GameStatistics returns general game statistic information on
// a per opponent basis.
// Requires you to be logged in.
func (c *Client) GameStatistics() ([]GameStatistic, error) {
	msg, err := c.makeRequest("/stats/my_game_stats", nil)
	if err == nil {
		return msg.GameStatistics, nil
	}

	return nil, err
}

// TopWriters gets the list of users that have submitted the
// most questions, that have also been accepted.
// Requires you to be logged in.
func (c *Client) TopWriters() ([]User, error) {
	msg, err := c.makeRequest("/users/top_list_writers", nil)
	if err == nil {
		return msg.Users, nil
	}

	return nil, err
}

// TopPlayers gets the list of users that have the highest ranking
// based on the points won in games.
// Requires you to be logged in.
func (c *Client) TopPlayers() ([]User, error) {
	msg, err := c.makeRequest("/users/top_list_rating", nil)
	if err == nil {
		return msg.Users, nil
	}

	return nil, err
}

// CategoryList fetches all possible categories.
// Requires you to be logged in.
func (c *Client) CategoryList() (map[int]string, error) {
	msg, err := c.makeRequest("/web/cats", nil)
	if err == nil {
		return msg.Categories, nil
	}

	return nil, err
}

// CategoryStatistics gives you the performance of the logged in
// user for all categories.
// Requires you to be logged in.
func (c *Client) CategoryStatistics() (*UserCategoryStatistics, error) {
	msg, err := c.makeRequest("/stats/my_stats", nil)
	if err == nil {
		return msg.UserCategoryStatistics, nil
	}

	return nil, err
}

func (c *Client) makeRequest(path string, data url.Values) (*message, error) {
	requestURL := protocolPrefix + hostName + path

	request, err := buildRequest(requestURL, data)

	if err != nil {
		return nil, err
	}

	clientDate := time.Now().Format("2006-01-02 15:04:05")

	request.Header.Set("dt", "a")
	request.Header.Set("authorization", getAuthCode(path, clientDate, data))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("User-Agent", userAgent)
	request.Header.Set("clientdate", clientDate)
	request.Header.Set("Accept-Encoding", "identity")

	// Got to load in the cookie manually and swap
	// the underscores to backslashes which Go doesn't
	// support natively in cookie values.
	cookies := c.Jar.Cookies(request.URL)
	if len(cookies) > 0 {
		for _, cookie := range cookies {
			s := cookie.Name + "=\"" + cookie.Value + "\""
			s = strings.Replace(s, "_", "\\", -1)

			if c := request.Header.Get("Cookie"); c != "" {
				request.Header.Set("Cookie", c+"; "+s)
			} else {
				request.Header.Set("Cookie", s)
			}
		}
	}

	resp, err := c.client.Do(request)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	// Using this little trick to save the cookies
	// manually, see the comment above.
	cookie := resp.Header.Get("Set-Cookie")
	if cookie != "" {
		cookie = strings.Replace(cookie, "\\", "_", -1)
		resp.Header.Set("Set-Cookie", cookie)
		c.Jar.SetCookies(request.URL, resp.Cookies())
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var m message
	err = json.Unmarshal(body, &m)
	return &m, err
}

func getAuthCode(path, clientDate string, data url.Values) string {
	msg := protocolPrefix + hostName + path + clientDate

	vals := make([]string, len(data))
	for _, v := range data {
		vals = append(vals, v[0])
	}
	sort.Strings(vals)
	msg += strings.Join(vals, "")

	h := hmac.New(sha256.New, []byte(authKey))
	io.WriteString(h, msg)
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func stringifyIntSlice(slice []int) string {
	l := len(slice) - 1
	s := "["
	for i, a := range slice {
		s += strconv.Itoa(a)
		if i != l {
			s += ", "
		}
	}
	s += "]"

	return s
}

func buildRequest(requestURL string, data url.Values, method ...string) (*http.Request, error) {
	if len(method) != 0 {
		if data == nil {
			return http.NewRequest(method[0], requestURL, nil)
		}

		if len(method) > 1 {
			return http.NewRequest(method[0], requestURL, data.Get(method[1]))
		}
		return http.NewRequest(method[0], requestURL, strings.NewReader(data.Encode()))
	}

	if data == nil {
		return http.NewRequest("GET", requestURL, nil)
	}
	return http.NewRequest("POST", requestURL, strings.NewReader(data.Encode()))
}
