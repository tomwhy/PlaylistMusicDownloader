package auth

type clientSecret struct {
	ClientID     string   `json:"client_id"`
	AuthURI      string   `json:"auth_uri"`
	TokenURI     string   `json:"token_uri"`
	ClientSecret string   `json:"client_secret"`
	RedirectURIS []string `json:"redirect_uris"`
}
