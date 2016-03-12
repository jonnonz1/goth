// Package slack implements the OAuth2 protocol for authenticating users through slack.
// This package can be used as a reference implementation of an OAuth2 provider for Goth.
package slack

import (
	"bytes"
	"encoding/json"
	"github.com/markbates/goth"
	"golang.org/x/oauth2"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const (
	authURL         string = "https://slack.com/oauth/authorize"
	tokenURL        string = "https://slack.com/api/oauth.access"
	endpointUser    string = "https://slack.com/api/auth.test"
	endpointProfile string = "https://slack.com/api/users.info"
)

// Provider is the implementation of `goth.Provider` for accessing Slack.
type Provider struct {
	ClientKey   string
	Secret      string
	CallbackURL string
	config      *oauth2.Config
}

// New creates a new Slack provider and sets up important connection details.
// You should always call `slack.New` to get a new provider.  Never try to
// create one manually.
func New(clientKey, secret, callbackURL string, scopes ...string) *Provider {
	p := &Provider{
		ClientKey:   clientKey,
		Secret:      secret,
		CallbackURL: callbackURL,
	}
	p.config = newConfig(p, scopes)
	return p
}

// Name is the name used to retrieve this provider later.
func (p *Provider) Name() string {
	return "slack"
}

// Debug is a no-op for the slack package.
func (p *Provider) Debug(debug bool) {}

// BeginAuth asks Slack for an authentication end-point.
func (p *Provider) BeginAuth(state string) (goth.Session, error) {
	return &Session{
		AuthURL: p.config.AuthCodeURL(state),
	}, nil
}

// FetchUser will go to Slack and access basic information about the user.
func (p *Provider) FetchUser(session goth.Session) (goth.User, error) {
	sess := session.(*Session)
	user := goth.User{
		AccessToken:  sess.AccessToken,
		Provider:     p.Name(),
		RefreshToken: sess.RefreshToken,
		ExpiresAt:    sess.ExpiresAt,
	}
	// Get the userID, slack needs userID in order to get user profile info
	response, err := http.Get(endpointUser + "?token=" + url.QueryEscape(sess.AccessToken))
	if err != nil {
		if response != nil {
			response.Body.Close()
		}
		return user, err
	}

	bits, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return user, err
	}

	u := struct {
		UserID string `json:"user_id"`
	}{}

	err = json.NewDecoder(bytes.NewReader(bits)).Decode(&u)

	// Get user profile info
	response, err = http.Get(endpointProfile + "?token=" + url.QueryEscape(sess.AccessToken) + "&user=" + u.UserID)
	if err != nil {
		if response != nil {
			response.Body.Close()
		}
		return user, err
	}
	defer response.Body.Close()

	bits, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return user, err
	}

	err = json.NewDecoder(bytes.NewReader(bits)).Decode(&user.RawData)
	if err != nil {
		return user, err
	}

	err = userFromReader(bytes.NewReader(bits), &user)
	return user, err
}

// UnmarshalSession wil unmarshal a JSON string into a session.
func (p *Provider) UnmarshalSession(data string) (goth.Session, error) {
	s := &Session{}
	err := json.NewDecoder(strings.NewReader(data)).Decode(s)
	return s, err
}

func newConfig(provider *Provider, scopes []string) *oauth2.Config {
	c := &oauth2.Config{
		ClientID:     provider.ClientKey,
		ClientSecret: provider.Secret,
		RedirectURL:  provider.CallbackURL,
		Endpoint: oauth2.Endpoint{
			AuthURL:  authURL,
			TokenURL: tokenURL,
		},
		Scopes: []string{},
	}

	if len(scopes) > 0 {
		for _, scope := range scopes {
			c.Scopes = append(c.Scopes, scope)
		}
	} else {
		c.Scopes = append(c.Scopes, "users:read")
	}
	return c
}

func userFromReader(r io.Reader, user *goth.User) error {
	u := struct {
		User struct {
			NickName string `json:"name"`
			ID       string `json:"id"`
			Profile  struct {
				Email     string `json:"email"`
				Name      string `json:"real_name"`
				AvatarURL string `json:"image_32"`
			} `json:"profile"`
		} `json:"user"`
	}{}
	err := json.NewDecoder(r).Decode(&u)
	if err != nil {
		return err
	}
	user.Email = u.User.Profile.Email
	user.Name = u.User.Profile.Name
	user.NickName = u.User.NickName
	user.UserID = u.User.ID
	user.AvatarURL = u.User.Profile.AvatarURL
	return nil
}

//RefreshTokenAvailable refresh token is provided by auth provider or not
func (p *Provider) RefreshTokenAvailable() bool {
	return false
}

//RefreshToken get new access token based on the refresh token
func (p *Provider) RefreshToken(refreshToken string) (*oauth2.Token, error) {
	return nil, nil
}