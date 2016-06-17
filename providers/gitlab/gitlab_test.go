package gitlab_test

import (
	"github.com/jonnonz1/goth"
	"github.com/jonnonz1/goth/providers/gitlab"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func Test_New(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	p := provider()

	a.Equal(p.ClientKey, os.Getenv("GITLAB_KEY"))
	a.Equal(p.Secret, os.Getenv("GITLAB_SECRET"))
	a.Equal(p.CallbackURL, "/foo")
}

func Test_Implements_Provider(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	a.Implements((*goth.Provider)(nil), provider())
}

func Test_BeginAuth(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	p := provider()
	session, err := p.BeginAuth("test_state")
	s := session.(*gitlab.Session)
	a.NoError(err)
	a.Contains(s.AuthURL, "gitlab.com/oauth/authorize")
}

func Test_SessionFromJSON(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	p := provider()
	session, err := p.UnmarshalSession(`{"AuthURL":"https://gitlab.com/oauth/authorize","AccessToken":"1234567890"}`)
	a.NoError(err)

	s := session.(*gitlab.Session)
	a.Equal(s.AuthURL, "https://gitlab.com/oauth/authorize")
	a.Equal(s.AccessToken, "1234567890")
}

func provider() *gitlab.Provider {
	return gitlab.New(os.Getenv("GITLAB_KEY"), os.Getenv("GITLAB_SECRET"), "/foo")
}
