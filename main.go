package main

import (
	"flag"
	"log"
	"os"

	"github.com/hashicorp/terraform/plugin"
	"github.com/hashicorp/terraform/terraform"
	"github.com/sourcegraph/terraform-provider-site24x7/site24x7"
	"github.com/sourcegraph/terraform-provider-site24x7/site24x7/oauth"
)

var oauthFile = flag.String("oauth-file", "", "(required) path to the oauth_file, will be created if it doesn't exist")

func main() {
	flag.Parse()

	if *oauthFile == "" {
		flag.PrintDefaults()
		os.Exit(2)
	}

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
