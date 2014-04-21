package quizduell

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
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
}

func NewClient(cookieJar http.CookieJar) *Client {
	if cookieJar == nil {
		cookieJar, _ = cookiejar.New(nil)
	}

	return &Client{
		client: &http.Client{
			Jar: cookieJar,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		},
	}
}

func (c *Client) Login(username, password string) {
	data := url.Values{}

	h := md5.New()
	io.WriteString(h, passwordSalt+password)
	data.Set("pwd", string(hex.EncodeToString(h.Sum(nil))))
	data.Set("name", username)

	fmt.Println(c.makeRequest("/users/login", data))
}

func (c *Client) makeRequest(path string, data url.Values) map[string]interface{} {
	requestURL := fmt.Sprintf("https://%s%s", hostName, path)

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
	json.Unmarshal(body, m)
	return m
}

func getAuthCode(path, clientDate string, data url.Values) string {
	msg := fmt.Sprintf("https://%s%s%s", hostName, path, clientDate)

	vals := make([]string, len(data))
	for _, v := range data {
		vals = append(vals, v[0])
	}
	sort.Strings(vals)
	msg += strings.Join(vals, "")
	fmt.Println(msg)

	h := hmac.New(sha256.New, []byte(authKey))
	io.WriteString(h, msg)
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
