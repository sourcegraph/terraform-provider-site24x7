package oauth

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"os"
	"sync"
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
	// Protects tkns.AccessToken (updated in Refresh goroutine)
	sync.RWMutex
	tkns      *tokens
	storePath string
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

	ator.Lock()
	ator.tkns.AccessToken, _ = result["access_token"].(string)
	ator.Unlock()

	ator.tkns.ExpiresInSec, _ = result["expires_in_sec"].(float64)

	refreshToken, _ := result["refresh_token"].(string)
	if refreshToken != "" {
		ator.tkns.RefreshToken = refreshToken
	}

	if ator.tkns.AccessToken == "" || ator.tkns.ExpiresInSec == 0 {
		return errors.New("error while fetching access token: empty token or no expiration")
	}
	ator.tkns.TokenGenerationTime = time.Now().UnixNano() / 1e6
	return ator.tkns.persist(ator.storePath)
}

func (ator *Authenticator) getAccessTokenFromRefreshToken() error {
	return ator.getAccessTokenFrom(getURL(ator.tkns.refreshTokenURLValues()))
}

func (ator *Authenticator) setAccessTokenFromCode() error {
	return ator.getAccessTokenFrom(getURL(ator.tkns.generatedCodeURLValues()))
}

func fileExists(path string) (bool, error) {
	fi, err := os.Lstat(path)
	if err == nil {
		if fi.IsDir() {
			return false, errors.New(path + " is a directory")
		}
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (ator *Authenticator) Refresh() error {
	if ator.tkns.AccessToken == "" {
		return errors.New("no access token to Refresh")
	}
	err := ator.getAccessTokenFromRefreshToken()
	if err != nil {
		// if we failed to Refresh we want to try again in 5 min
		ator.tkns.ExpiresInSec = 300
		return err
	}
	return nil
}

func (ator *Authenticator) scheduleRefresh() {
	go func() {
		for {
			timer := time.NewTimer(time.Second * time.Duration(math.Max(ator.tkns.ExpiresInSec-10, 2)))
			<-timer.C
			err := ator.Refresh()
			if err != nil {
				fmt.Printf("failed to Refresh access token: %b", err)
			}
		}
	}()
}

// AccessToken returns the access token.
func (ator *Authenticator) AccessToken() string {
	ator.RLock()
	defer ator.RUnlock()

	return ator.tkns.AccessToken
}

// NewAuthenticator creates an authenticator that will acquire an access token by reading it from the specified
// file if it exists or by requesting it from Zoho and saving it to the specified file. It also schedules a
// Refresh goroutine which will Refresh the token ten seconds before it expires.
func NewAuthenticator(path string) (*Authenticator, error) {
	ator := &Authenticator{
		storePath: path,
	}
	storeExists, err := fileExists(ator.storePath)
	if err != nil {
		return nil, err
	}

	if !storeExists {
		// persist a template so humans can add client id, client secret and generated code
		var tkns tokens
		err = tkns.persist(ator.storePath)
		if err != nil {
			return nil, err
		}
		return nil, errors.New("update CLIENT_ID, CLIENT_SECRET and GENERATED_CODE in the file " + ator.storePath)
	}

	tkns, err := load(ator.storePath)
	if err != nil {
		return nil, err
	}
	ator.tkns = tkns

	if ator.tkns.AccessToken != "" {
		ator.scheduleRefresh()
		return ator, nil
	}

	if ator.tkns.ClientId == "" || ator.tkns.ClientSecret == "" || ator.tkns.GeneratedCode == "" {
		return nil, errors.New("update CLIENT_ID, CLIENT_SECRET and GENERATED_CODE in the file " + ator.storePath)
	}
	err = ator.setAccessTokenFromCode()
	if err != nil {
		return nil, err
	}
	ator.scheduleRefresh()
	return ator, nil
}
