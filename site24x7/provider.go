package site24x7

import (
	"net/http"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"authtoken": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("SITE24X7_AUTHTOKEN", nil),
				Description: "Authorization Token for Site24x7.",
			},
		},
		DataSourcesMap: map[string]*schema.Resource{
			"site24x7_ip_whitelist": dataSourceSite24x7IpWhitelist(),
		},

		ResourcesMap: map[string]*schema.Resource{
			"site24x7_website_monitor": resourceSite24x7WebsiteMonitor(),
		},

		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	h := make(http.Header)
	h.Set("Authorization", "Zoho-authtoken "+d.Get("authtoken").(string))
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
