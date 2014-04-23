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

func (c *Client) Login(username, password string) map[string]interface{} {
	data := url.Values{}

	h := md5.New()
	io.WriteString(h, passwordSalt+password)
	data.Set("pwd", string(hex.EncodeToString(h.Sum(nil))))
	data.Set("name", username)

	return c.makeRequest("/users/login", data)
}

func (c *Client) CreateUser(username, email, password string) map[string]interface{} {
	data := url.Values{}

	data.Set("name", username)
	if email != "" {
		data.Set("email", email)
	}

	h := md5.New()
	io.WriteString(h, passwordSalt+password)
	data.Set("pwd", string(hex.EncodeToString(h.Sum(nil))))

	return c.makeRequest("/users/create", data)
}

func (c *Client) UpdateUser(username, email, password string) map[string]interface{} {
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

	return c.makeRequest("/users/update_user", data)
}

func (c *Client) FindUser(username string) map[string]interface{} {
	data := url.Values{}
	data.Set("opponent_name", username)

	return c.makeRequest("/users/find_user", data)
}

func (c *Client) AddFriend(userID string) map[string]interface{} {
	data := url.Values{}

	data.Set("friend_id", userID)

	return c.makeRequest("/users/add_friend", data)
}

func (c *Client) RemoveFriend(userID string) map[string]interface{} {
	data := url.Values{}

	data.Set("friend_id", userID)

	return c.makeRequest("/users/remove_friend", data)
}

func (c *Client) UpdateAvatar(avatarCode string) map[string]interface{} {
	data := url.Values{}

	data.Set("avatar_code", avatarCode)

	return c.makeRequest("/users/update_avatar", data)
}

// Not quite sure how this functionality works
func (c *Client) SendForgotPasswordEmail(email string) map[string]interface{} {
	data := url.Values{}

	data.Set("email", email)

	return c.makeRequest("/users/forgot_pwd", data)
}

func (c *Client) AddBlocked(userID string) map[string]interface{} {
	data := url.Values{}

	data.Set("blocked_id", userID)

	return c.makeRequest("/users/add_blocked", data)
}

func (c *Client) RemoveBlocked(userID string) map[string]interface{} {
	data := url.Values{}

	data.Set("blocked_id", userID)

	return c.makeRequest("/users/remove_blocked", data)
}

func (c *Client) StartGame(opponentID string) map[string]interface{} {
	data := url.Values{}

	data.Set("opponent_id", opponentID)

	return c.makeRequest("/games/create_game", data)
}

func (c *Client) StartRandomGame() map[string]interface{} {
	return c.makeRequest("/games/start_random_game", nil)
}

func (c *Client) GetGame(gameID string) map[string]interface{} {
	return c.makeRequest("/games/"+gameID, nil)
}

func (c *Client) GiveUp(gameID string) map[string]interface{} {
	data := url.Values{}

	data.Set("game_id", gameID)

	return c.makeRequest("/games/give_up", data)
}

func (c *Client) AcceptGame(gameID string) map[string]interface{} {
	data := url.Values{}

	data.Set("accept", "1")
	data.Set("game_id", gameID)
	return c.makeRequest("/games/accept", data)
}

func (c *Client) DeclineGame(gameID string) map[string]interface{} {
	data := url.Values{}

	data.Set("accept", "0")
	data.Set("game_id", gameID)
	return c.makeRequest("/games/accept", data)
}

func (c *Client) UploadRoundAnswers(gameID string, answers []int, categoryID string) map[string]interface{} {
	data := url.Values{}

	data.Set("game_id", gameID)
	for _, a := range answers {
		data.Add("answers", string(a))
	}
	data.Set("cat_choice", categoryID)

	return c.makeRequest("/games/upload_round_answers", data)
}

func (c *Client) GetUserGames() map[string]interface{} {
	return c.makeRequest("/users/current_user_games", url.Values{})
}

func (c *Client) SendMessage(gameID, message string) map[string]interface{} {
	data := url.Values{}

	data.Set("game_id", gameID)
	data.Set("text", message)

	return c.makeRequest("/games/send_message", data)
}

func (c *Client) GameStatistics() map[string]interface{} {
	return c.makeRequest("/stats/my_game_stats", nil)
}

func (c *Client) TopWriters() map[string]interface{} {
	return c.makeRequest("/users/top_list_writers", nil)
}

func (c *Client) TopPlayers() map[string]interface{} {
	return c.makeRequest("/users/top_list_rating", nil)
}

func (c *Client) CategoryList() map[string]interface{} {
	return c.makeRequest("/web/cats", nil)
}

func (c *Client) CategoryStatistics() map[string]interface{} {
	return c.makeRequest("/stats/my_stats", nil)
}

func (c *Client) makeRequest(path string, data url.Values) map[string]interface{} {
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

	var m map[string]interface{}
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
