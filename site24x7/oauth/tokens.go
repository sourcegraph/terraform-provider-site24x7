package oauth

import (
	"net/url"
	"time"
)

type tokens struct {
	ClientId            string  `json:"CLIENT_ID"`
	ClientSecret        string  `json:"CLIENT_SECRET"`
	GeneratedCode       string  `json:"GENERATED_CODE"`
	RefreshToken        string  `json:"REFRESH_TOKEN"`
	AccessToken         string  `json:"ACCESS_TOKEN"`
	TokenGenerationTime int64   `json:"TOKEN_GENERATION_TIME"`
	ExpiresInSec        float64 `json:"EXPIRES_IN_SEC"`
}

func (tkns *tokens) refreshTokenURLValues() url.Values {
	urlValues := url.Values{}
	urlValues.Set("client_id", tkns.ClientId)
	urlValues.Set("client_secret", tkns.ClientSecret)
	urlValues.Set("refresh_token", tkns.RefreshToken)
	urlValues.Set("grant_type", "refresh_token")
	return urlValues
}

func (tkns *tokens) generatedCodeURLValues() url.Values {
	urlValues := url.Values{}
	urlValues.Set("client_id", tkns.ClientId)
	urlValues.Set("client_secret", tkns.ClientSecret)
	urlValues.Set("code", tkns.GeneratedCode)
	urlValues.Set("grant_type", "authorization_code")
	return urlValues
}

func (tkns *tokens) expired() bool {
	now := time.Now().UnixNano() / 1e6
	return now-tkns.TokenGenerationTime > int64(tkns.ExpiresInSec)*1000
}
