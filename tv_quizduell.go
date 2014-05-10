package quizduell

import (
	"net/http"
	"net/url"
)

const (
	tvProtocolPrefix = "https://"
	tvHostName       = "quizduell.mobilemassresponse.de"
	corsHeaderToken  = "grandc3ntr1xrul3z"
)

type TVClient struct {
	UserID int
	// The API seems to be using this auth token (tt)
	// as a validation mechanism, instead of cookies
	// or the like.
	AuthToken string
}

// NewTVClient creates a new TV client that can be used
// to interact with the TV version of Quizduell.
// The authToken is User.TT
func NewTVClient(userID int, authToken string) *TVClient {
	return &TVClient{
		UserID:    userID,
		AuthToken: authToken,
	}
}

// FromClient returns a new TV client based on an already
// existant (and logged in) Quizduell client. If the user
// hasn't created a TV profile yet, this will also be done
// in the process.
func FromClient(c *Client) *TVClient {
	user := c.CreateTVUser()
	return NewTVClient(user.UserID, user.TT)
}

func (t *TVClient) AgreeAGBs() map[string]interface{} {
	return t.request("/feousers/agbs/"+strconv.Itoa(t.UserID)+"/true", url.Values{})
}

func (t *TVClient) request(path string, data url.Values) map[string]interface{} {
	requestURL := tvProtocolPrefix + tvHostName + path
	request, err := buildRequest(requestURL, data)

	if err != nil {
		panic(err)
	}

	request.Header.Set("x-app-request", corsHeaderToken)
	request.Header.Set("x-tv-authtoken", t.AuthToken)

	resp, err := http.DefaultClient.Do(request)

	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		panic(err)
	}

	var m map[string]interface{}
	err = json.Unmarshal(body, &m)
	return m
}
