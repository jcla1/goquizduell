package quizduell

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
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
func FromClient(c *Client) (*TVClient, error) {
	user, err := c.CreateTVUser()
	if err != nil {
		return nil, err
	}

	return NewTVClient(user.ID, user.TT), nil
}

// AgreeAGBs makes the current user agree to the AGB
// put up by the TV quiz broadcaster.
func (t *TVClient) AgreeAGBs() (map[string]interface{}, error) {
	return t.request("/feousers/agbs/"+strconv.Itoa(t.UserID)+"/true", url.Values{})
}

// GetState returns the state of the TV quiz
func (t *TVClient) GetState() (map[string]interface{}, error) {
	return t.request("/states/"+strconv.Itoa(t.UserID), nil)
}

func (t *TVClient) GetRankings() (map[string]interface{}, error) {
	return t.request("/users/myranking/"+strconv.Itoa(t.UserID), nil)
}

func (t *TVClient) GetMyProfile() (map[string]interface{}, error) {
	return t.GetProfile(t.UserID)
}

func (t *TVClient) GetProfile(userID int) (map[string]interface{}, error) {
	return t.request("/users/profiles/"+strconv.Itoa(userID), nil)
}

func (t *TVClient) PostProfile(profile map[string]string) (map[string]interface{}, error) {
	data := url.Values{}

	for key, val := range profile {
		data.Set(key, val)
	}

	return t.request("/users/profiles/"+strconv.Itoa(t.UserID), data)
}

func (t *TVClient) DeleteUser() (map[string]interface{}, error) {
	return t.request("/users/profiles/"+strconv.Itoa(t.UserID), nil, "DELETE")
}

func (t *TVClient) SetAvatarAndNickname(nick, avatarCode string) (map[string]interface{}, error) {
	data := url.Values{}

	if avatarCode != "" {
		data.Set("AvatarString", avatarCode)
	}
	data.Set("Nick", nick)

	return t.request("/users/"+strconv.Itoa(t.UserID)+"/avatarandnick", data)
}

func (t *TVClient) SelectCategory(categoryID int) (map[string]interface{}, error) {
	return t.request("/users/"+strconv.Itoa(t.UserID)+"/category"+strconv.Itoa(categoryID), nil)
}

func (t *TVClient) SendAnswer(questionID, answerID int) (map[string]interface{}, error) {
	return t.request("/users/"+strconv.Itoa(t.UserID)+"/response"+strconv.Itoa(questionID)+"/"+strconv.Itoa(answerID), nil)
}

func (t *TVClient) UploadProfileImage(r io.Reader) (map[string]interface{}, error) {
	img, _ := ioutil.ReadAll(r)

	data := url.Values{}
	data.Set("img", base64.StdEncoding.EncodeToString(img))

	return t.request("/users/base64/"+strconv.Itoa(t.UserID)+"/jpg", data, "POST", "img")
}

func (t *TVClient) request(path string, data url.Values, method ...string) (map[string]interface{}, error) {
	requestURL := tvProtocolPrefix + tvHostName + path
	request, err := buildRequest(requestURL, data, method...)

	if err != nil {
		return nil, err
	}

	request.Header.Set("x-app-request", corsHeaderToken)
	request.Header.Set("x-tv-authtoken", t.AuthToken)
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=utf-8")

	resp, err := http.DefaultClient.Do(request)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	var m map[string]interface{}
	err = json.Unmarshal(body, &m)

	if err != nil {
		return nil, err
	}

	return m, nil
}
