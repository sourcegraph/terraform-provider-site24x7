package main

import (
	"flag"
	"log"

	"github.com/hashicorp/terraform/plugin"
	"github.com/hashicorp/terraform/terraform"
	"github.com/sourcegraph/terraform-provider-site24x7/site24x7"
	"github.com/sourcegraph/terraform-provider-site24x7/site24x7/oauth"
)

var oauthFile = flag.String("oauth_file", "", "path to the oauth_file")

func main() {
	flag.Parse()

	ator, err := oauth.NewAuthenticator(*oauthFile)
	if err != nil {
		log.Fatal(err)
	}

	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: func() terraform.ResourceProvider {
			return site24x7.Provider(ator)
		},
	})
}
