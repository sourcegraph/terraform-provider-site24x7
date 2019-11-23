package site24x7

import (
	"net/http"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"github.com/sourcegraph/terraform-provider-site24x7/site24x7/oauth"
)

func Provider(ator *oauth.Authenticator) terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"authtoken": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: func() (i interface{}, err error) {
					accessToken := ator.AccessToken()
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
