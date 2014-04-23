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
	hostName     = "qkgermany.feomedia.se"
	userAgent    = "Quizduell A 1.3.2"
	authKey      = "irETGpoJjG57rrSC"
	passwordSalt = "SQ2zgOTmQc8KXmBP"
	timeout      = 20000
)

type Client struct {
	client *http.Client
	// We're separating out this cookie jar from
	// the HTTP client, because Go is trying to be
	// peticularly RFC compliant and doesn't allow
	// certain characters in cookies. That's why we
	// have to handle them seperately.
	jar http.CookieJar
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
		jar: cookieJar,
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

	if name != "" {
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

func (c *Client) AddFriend(userId string) map[string]interface{} {
	data := url.Values{}

	data.Set("friend_id", userId)

	return c.makeRequest("/users/add_friend", data)
}

func (c *Client) RemoveFriend(userId string) map[string]interface{} {
	data := url.Values{}

	data.Set("friend_id", userId)

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

func (c *Client) AddBlocked(userId string) map[string]interface{} {
	data := url.Values{}

	data.Set("blocked_id", userId)

	return c.makeRequest("/users/add_blocked", data)
}

func (c *Client) RemoveBlocked(userId string) map[string]interface{} {
	data := url.Values{}

	data.Set("blocked_id", userId)

	return c.makeRequest("/users/remove_blocked", data)
}

func (c *Client) TopWriters() map[string]interface{} {
	return c.makeRequest("/users/top_list_writers", url.Values{})
}

func (c *Client) TopPlayers() map[string]interface{} {
	return c.makeRequest("/users/top_list_rating", url.Values{})
}

func (c *Client) makeRequest(path string, data url.Values) map[string]interface{} {
	requestURL := "https://" + hostName + path

	request, err := http.NewRequest("POST", requestURL, strings.NewReader(data.Encode()))
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

	resp, err := c.client.Do(request)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	var m map[string]interface{}
	json.Unmarshal(body, &m)
	return m
}

func getAuthCode(path, clientDate string, data url.Values) string {
	msg := "https://" + hostName + path + clientDate

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
