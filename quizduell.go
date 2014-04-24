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
	timeout        = 20000
)

type Client struct {
	client *http.Client
	// We're separating out this cookie jar from
	// the HTTP client, because Go is trying to be
	// peticularly RFC compliant and doesn't allow
	// certain characters in cookies. That's why we
	// have to handle them seperately.
	Jar http.CookieJar
}

func NewClient(cookieJar http.CookieJar) *Client {
	if cookieJar == nil {
		jar, err := cookiejar.New(nil)
		if err != nil {
			panic(err)
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
	}
}

func (c *Client) Login(username, password string) *Status {
	data := url.Values{}

	h := md5.New()
	io.WriteString(h, passwordSalt+password)
	data.Set("pwd", string(hex.EncodeToString(h.Sum(nil))))
	data.Set("name", username)

	return c.makeRequest("/users/login", data).Status
}

func (c *Client) CreateUser(username, email, password string) *Status {
	data := url.Values{}

	data.Set("name", username)
	if email != "" {
		data.Set("email", email)
	}

	h := md5.New()
	io.WriteString(h, passwordSalt+password)
	data.Set("pwd", string(hex.EncodeToString(h.Sum(nil))))

	return c.makeRequest("/users/create", data).Status
}

func (c *Client) UpdateUser(username, email, password string) *Status {
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

	return c.makeRequest("/users/update_user", data).Status
}

func (c *Client) FindUser(username string) *User {
	data := url.Values{}
	data.Set("opponent_name", username)

	return c.makeRequest("/users/find_user", data).U
}

func (c *Client) AddFriend(userID int) *Popup {
	data := url.Values{}

	data.Set("friend_id", strconv.Itoa(userID))

	return c.makeRequest("/users/add_friend", data).Popup
}

func (c *Client) RemoveFriend(userID int) *Popup {
	data := url.Values{}

	data.Set("friend_id", strconv.Itoa(userID))

	return c.makeRequest("/users/remove_friend", data).Popup
}

func (c *Client) UpdateAvatar(avatarCode string) bool {
	data := url.Values{}

	data.Set("avatar_code", avatarCode)

	return c.makeRequest("/users/update_avatar", data).T
}

// Not quite sure how this functionality works
func (c *Client) SendForgotPasswordEmail(email string) *Popup {
	data := url.Values{}

	data.Set("email", email)

	return c.makeRequest("/users/forgot_pwd", data).Popup
}

func (c *Client) AddBlocked(userID int) []User {
	data := url.Values{}

	data.Set("blocked_id", strconv.Itoa(userID))

	return c.makeRequest("/users/add_blocked", data).Blocked
}

func (c *Client) RemoveBlocked(userID int) []User {
	data := url.Values{}

	data.Set("blocked_id", strconv.Itoa(userID))

	return c.makeRequest("/users/remove_blocked", data).Blocked
}

func (c *Client) StartGame(opponentID int) *Game {
	data := url.Values{}

	data.Set("opponent_id", strconv.Itoa(opponentID))

	return c.makeRequest("/games/create_game", data).Game
}

func (c *Client) StartRandomGame() *Game {
	return c.makeRequest("/games/start_random_game", nil).Game
}

func (c *Client) GetGame(gameID int) *Game {
	return c.makeRequest("/games/"+strconv.Itoa(gameID), nil).Game
}

func (c *Client) GiveUp(gameID int) (*Game, *Popup) {
	data := url.Values{}

	data.Set("game_id", strconv.Itoa(gameID))

	d := c.makeRequest("/games/give_up", data)
	return d.Game, d.Popup
}

func (c *Client) AcceptGame(gameID int) bool {
	data := url.Values{}

	data.Set("accept", "1")
	data.Set("game_id", strconv.Itoa(gameID))
	return c.makeRequest("/games/accept", data).T
}

func (c *Client) DeclineGame(gameID int) bool {
	data := url.Values{}

	data.Set("accept", "0")
	data.Set("game_id", strconv.Itoa(gameID))
	return c.makeRequest("/games/accept", data).T
}

func (c *Client) UploadRoundAnswers(gameID int, answers []int, categoryID int) *Game {
	data := url.Values{}

	l := len(answers) - 1
	s := "["
	for i, a := range answers {
		s += strconv.Itoa(a)
		if i != l {
			s += ", "
		}
	}
	s += "]"

	data.Set("game_id", strconv.Itoa(gameID))
	data.Set("cat_choice", strconv.Itoa(categoryID))
	data.Set("answers", s)

	return c.makeRequest("/games/upload_round_answers", data).Game
}

func (c *Client) GetUserGames() *Status {
	return c.makeRequest("/users/current_user_games", url.Values{}).Status
}

func (c *Client) SendMessage(gameID int, message string) *InGameMessage {
	data := url.Values{}

	data.Set("game_id", strconv.Itoa(gameID))
	data.Set("text", message)

	return c.makeRequest("/games/send_message", data).InGameMessage
}

func (c *Client) GameStatistics() []GameStatistic {
	return c.makeRequest("/stats/my_game_stats", nil).GameStatistics
}

func (c *Client) TopWriters() []User {
	return c.makeRequest("/users/top_list_writers", nil).Users
}

func (c *Client) TopPlayers() []User {
	return c.makeRequest("/users/top_list_rating", nil).Users
}

func (c *Client) CategoryList() map[int]string {
	return c.makeRequest("/web/cats", nil).Categories
}

func (c *Client) CategoryStatistics() *UserCategoryStatistics {
	return c.makeRequest("/stats/my_stats", nil).UserCategoryStatistics
}

func (c *Client) makeRequest(path string, data url.Values) message {
	requestURL := protocolPrefix + hostName + path

	var request *http.Request
	var err error

	if data == nil {
		request, err = http.NewRequest("GET", requestURL, nil)
	} else {
		request, err = http.NewRequest("POST", requestURL, strings.NewReader(data.Encode()))
	}

	if err != nil {
		panic(err)
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
		panic(err)
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

	body, _ := ioutil.ReadAll(resp.Body)

	var m message
	json.Unmarshal(body, &m)
	return m
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
