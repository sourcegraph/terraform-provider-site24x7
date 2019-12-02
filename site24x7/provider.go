package site24x7

import (
	"net/http"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"github.com/sourcegraph/terraform-provider-site24x7/site24x7/oauth"
)

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"oauth_client_id": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("SITE24X7_CLIENT_ID", nil),
				Description: "Zoho Site24x7 OAuth2 client id.",
			},
			"oauth_client_secret": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("SITE24X7_CLIENT_SECRET", nil),
				Description: "Zoho Site24x7 OAuth2 client secret.",
			},
			"oauth_refresh_token": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("SITE24X7_REFRESH_TOKEN", nil),
				Description: "Zoho Site24x7 OAuth2 refresh token.",
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"site24x7_website_monitor": resourceSite24x7WebsiteMonitor(),
		},

		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	clientId := d.Get("oauth_client_id").(string)
	clientSecret := d.Get("oauth_client_secret").(string)
	refreshToken := d.Get("oauth_refresh_token").(string)

	ator, err := oauth.NewAuthenticator(clientId, clientSecret, refreshToken)
	if err != nil {
		return nil, err
	}

	h := make(http.Header)
	h.Set("Authorization", "Zoho-oauthtoken "+ator.AccessToken())
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
