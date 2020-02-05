package oauth

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const (
	tokenDomain = "https://accounts.zoho.com"
	tokenApi    = "/oauth/v2/token"
	timeout     = time.Second
)

func setRequestHeaders(request *http.Request, accessToken string) {
	request.Header.Set("Content-Type", "application/json;charset=UTF-8")
	request.Header.Set("Accept", "application/json; version=2.0")
	request.Header.Set("Authorization", "Zoho-oauthtoken "+accessToken)
}

func getURL(urlValues url.Values) string {
	baseAccURL := fmt.Sprintf("%s%s", tokenDomain, tokenApi)
	urlToReturn := fmt.Sprintf("%s?%s", baseAccURL, urlValues.Encode())
	return urlToReturn
}

type Authenticator struct {
	tkns *tokens
}

func (ator *Authenticator) getClient() *http.Client {
	trTLS11 := &http.Transport{
		DisableKeepAlives: true,
		TLSClientConfig: &tls.Config{
			MaxVersion:         tls.VersionTLS13,
			InsecureSkipVerify: true,
		},
	}
	client := http.Client{
		Transport: trTLS11,
		Timeout:   timeout,
	}
	return &client
}

func (ator *Authenticator) getAccessTokenFrom(urlToPost string) error {
	var result map[string]interface{}
	req, err := http.NewRequest("POST", urlToPost, nil)
	if err != nil {
		return err
	}
	setRequestHeaders(req, ator.tkns.AccessToken)
	client := ator.getClient()

	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	err = json.NewDecoder(res.Body).Decode(&result)
	if err != nil {
		return err
	}

	respErr, _ := result["error"].(string)
	if respErr != "" {
		return errors.New(respErr)
	}

	ator.tkns.AccessToken, _ = result["access_token"].(string)
	ator.tkns.ExpiresInSec, _ = result["expires_in_sec"].(float64)

	refreshToken, _ := result["refresh_token"].(string)
	if refreshToken != "" {
		ator.tkns.RefreshToken = refreshToken
	}

	if ator.tkns.AccessToken == "" {
		return errors.New("error while fetching access token: empty token")
	}

	if ator.tkns.ExpiresInSec == 0 {
		return errors.New("error while fetching access token: no expiration")
	}
	ator.tkns.TokenGenerationTime = time.Now().UnixNano() / 1e6
	return nil
}

func (ator *Authenticator) getAccessTokenFromRefreshToken() error {
	return ator.getAccessTokenFrom(getURL(ator.tkns.refreshTokenURLValues()))
}

func (ator *Authenticator) setAccessTokenFromCode() error {
	return ator.getAccessTokenFrom(getURL(ator.tkns.generatedCodeURLValues()))
}

func (ator *Authenticator) refresh() error {
	err := ator.getAccessTokenFromRefreshToken()
	if err != nil {
		// if we failed to refresh we want to try again in 30 secs
		ator.tkns.ExpiresInSec = 30
		return err
	}
	return nil
}

// AccessToken returns the access token.
func (ator *Authenticator) AccessToken() string {
	return ator.tkns.AccessToken
}

// NewAuthenticator creates an authenticator that will acquire an access token from Zoho using the specified
// client id, client secret and refresh token.
// If an error occurred while obtaining the access token an error is returned.
func NewAuthenticator(clientId, clientSecret, refreshToken string) (*Authenticator, error) {
	ator := &Authenticator{
		tkns: &tokens{
			ClientId:     clientId,
			ClientSecret: clientSecret,
			RefreshToken: refreshToken,
		},
	}

	err := ator.refresh()
	if err != nil {
		return nil, err
	}

	return ator, nil
}

// GenerateRefreshToken returns a refresh token given the specified client id, client secret and generate code token.
// If acquiring the refresh token fails then it returns an error.
func GenerateRefreshToken(clientId, clientSecret, generateCode string) (string, error) {
	ator := &Authenticator{
		tkns: &tokens{},
	}
	ator.tkns.ClientId = clientId
	ator.tkns.ClientSecret = clientSecret
	ator.tkns.GeneratedCode = generateCode

	err := ator.setAccessTokenFromCode()

	if err != nil {
		return "", err
	}
	return ator.tkns.RefreshToken, nil
}
