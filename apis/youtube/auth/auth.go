package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net/http"

	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const stateLength = 128

type Authorizer struct {
	ctx  context.Context
	conf *oauth2.Config
}

func NewAuthorizer(clientID, clientSecret, redirectURI string, scopes []string) Authorizer {
	return Authorizer{context.Background(),
		&oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Scopes:       scopes,
			Endpoint:     google.Endpoint,
			RedirectURL:  redirectURI,
		},
	}
}

func (auth *Authorizer) GetAuthURL() (string, string, error) {

	state, err := createState()
	if err != nil {
		logrus.Error("Failed generating random state.", err)
	}

	return auth.conf.AuthCodeURL(state, oauth2.AccessTypeOffline), state, err
}

func (auth *Authorizer) GetToken(authCode string) (*oauth2.Token, error) {
	return auth.conf.Exchange(auth.ctx, authCode)
}

func (auth *Authorizer) CreateClient(token oauth2.Token) *http.Client {

	return auth.conf.Client(auth.ctx, &token)
}

func createState() (string, error) {
	res := make([]byte, stateLength)

	_, err := rand.Read(res)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(res), nil
}
