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
	request.Header.Set("Authorization", "Zoho-oauthtoken " + accessToken)
}

func getUrl(urlValues url.Values) string {
	baseAccurl := fmt.Sprintf("%s%s", tokenDomain, tokenApi)
	urlToReturn := fmt.Sprintf("%s?%s", baseAccurl, urlValues.Encode())
	return urlToReturn
}

type Authenticator struct {
	// Protects tkns.AccessToken (updated in refresh goroutine)
	sync.Mutex
	tkns      tokens
	storePath string
}

func (ator *Authenticator) getClient() *http.Client {
	trTls11 := &http.Transport{
		DisableKeepAlives: true,
		TLSClientConfig: &tls.Config{
			MaxVersion:         tls.VersionTLS13,
			InsecureSkipVerify: true,
		},
	}
	client := http.Client{
		Transport: trTls11,
		Timeout:   timeout,
	}
	return &client
}

func (ator *Authenticator) setDefaultHeaders(request *http.Request) {
	request.Header.Set("Accept", "application/x-www-form-urlencoded")
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
	ator.Lock()
	ator.tkns.AccessToken = result["access_token"].(string)
	ator.Unlock()

	ator.tkns.ExpiresInSec = result["expires_in_sec"].(float64)
	if ator.tkns.AccessToken == "" || ator.tkns.ExpiresInSec == 0 {
		return errors.New("error while fetching access token from refresh token: empty token or no expiration")
	}
	ator.tkns.TokenGenerationTime = time.Now().UnixNano() / 1e6
	return ator.tkns.persist(ator.storePath)
}

func (ator *Authenticator) getAccessTokenFromRefreshToken() error {
	return ator.getAccessTokenFrom(getUrl(ator.tkns.refreshTokenURLValues()))
}

func (ator *Authenticator) setAccessTokenFromCode() error {
	return ator.getAccessTokenFrom(getUrl(ator.tkns.generatedCodeURLValues()))
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

func (ator *Authenticator) refresh() error {
	if ator.tkns.AccessToken == "" {
		return errors.New("no access token to refresh")
	}
	if ator.tkns.expired() {
		return ator.getAccessTokenFromRefreshToken()
	}
	return nil
}

func (ator *Authenticator) scheduleRefresh() {
	go func() {
		for {
			timer := time.NewTimer(time.Second * time.Duration(math.Max(ator.tkns.ExpiresInSec - 10, 2)))
			<-timer.C
			err := ator.refresh()
			if err != nil {
				fmt.Printf("failed to refresh access token: %b", err)
			}
		}
	}()
}

// AccessToken returns the access token.
func (ator *Authenticator) AccessToken() string {
	ator.Lock()
	defer ator.Unlock()

	return ator.tkns.AccessToken
}

// NewAuthenticator creates an authenticator that will acquire an access token by reading it from the specified
// file if it exists or by requuesting it from Zoho and saving it to the specified file. It also schedules a
// refresh goroutine which will refresh the token ten seconds before it expires.
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
		err = ator.tkns.persist(ator.storePath)
		if err != nil {
			return nil, err
		}
		return nil, errors.New("update CLIENT_ID, CLIENT_SECRET and GENERATED_CODE in the file " + ator.storePath)
	}

	err = ator.tkns.load(ator.storePath)
	if err != nil {
		return nil, err
	}

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
