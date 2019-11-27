package site24x7

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"github.com/sourcegraph/terraform-provider-site24x7/site24x7/oauth"
)

func Provider() terraform.ResourceProvider {
	// these two are captured by the DefaultFunc closure below
	var authenticator *oauth.Authenticator
	var oauthErr error

	oauthFile, ok := os.LookupEnv("SITE24X7_AUTHTOKEN_FILE")
	if !ok {
		currentDir, err := os.Getwd()
		if err != nil {
			oauthErr = err
		}
		oauthFile = filepath.Join(currentDir, "site24x7-oauth.json")
	}

	if oauthErr == nil {
		ator, err := oauth.NewAuthenticator(oauthFile)
		if err != nil {
			oauthErr = err
		} else {
			authenticator = ator
		}
	}

	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"oauthtoken": {
				Type:     schema.TypeString,
				Required: true,
				DefaultFunc: func() (i interface{}, err error) {
					if oauthErr != nil {
						return nil, oauthErr
					}
					accessToken := authenticator.AccessToken()
					return accessToken, nil
				},
				Description: "Username for StatusCake Account.",
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"site24x7_website_monitor": resourceSite24x7WebsiteMonitor(),
		},

		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	h := make(http.Header)
	h.Set("Authorization", "Zoho-oauthtoken "+d.Get("oauthtoken").(string))
	return &http.Client{
		Transport: &staticHeaderTransport{
			base:   http.DefaultTransport,
			header: h,
		},
	}, nil
}

type staticHeaderTransport struct {
	base   http.RoundTripper
	header http.Header
}

func (t *staticHeaderTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	for k, v := range t.header {
		req.Header[k] = v
	}
	return t.base.RoundTrip(req)
}
