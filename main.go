package main

import (
	"github.com/hashicorp/terraform/plugin"
	"github.com/hashicorp/terraform/terraform"
	"github.com/sourcegraph/terraform-provider-site24x7/site24x7"
)

// main is a plugin main, not a "real" main.
// Please set the value of environment variable `SITE24X7_AUTHTOKEN_FILE` to a path to a JSDN file
// This JSDN file stores the client id, client secret and refresh token necessary to obtain OAuth2 access tokens from
// site24x7. Expected is the following content for the JSON file:
//
// {
//    "CLIENT_ID": "xxxx_your_client_id_xxxxx",
//    "CLIENT_SECRET": "xxxx_your_client_secret_xxxxx",
//    "REFRESH_TOKEN": "xxxx_your_refresh_token_xxxxx",
// }
//
// If `SITE24X7_AUTHTOKEN_FILE` is not set it will use `site24x7-oauth.json` in the current working directory.
// A helper command in site24x7/oauth/cmd/site24x7-oauth can be used to generate the JSON file.
func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: func() terraform.ResourceProvider {
			return site24x7.Provider()
		},
	})
}
