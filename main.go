package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/hashicorp/terraform/plugin"
	"github.com/hashicorp/terraform/terraform"
	"github.com/sourcegraph/terraform-provider-site24x7/site24x7"
	"github.com/sourcegraph/terraform-provider-site24x7/site24x7/oauth"
)

// main is a plugin main, not a "real" main.
// Please set the value of environment variable SITE24X7_AUTHTOKEN_FILE to a path to a JSDN file (file does not need to exist, it gets created if it doesn't).
// This JSDN file stores the OAuth2 tokens from site24x7. If you don't set this environment variable it will use path `.size24x7_auth.json`.
func main() {
	oauthFile, ok := os.LookupEnv("SITE24X7_AUTHTOKEN_FILE")
	if !ok {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatal(err)
		}
		oauthFile = filepath.Join(homeDir, ".size24x7_auth.json")
	}

	ator, err := oauth.NewAuthenticator(oauthFile)
	if err != nil {
		log.Fatal(err)
	}

	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: func() terraform.ResourceProvider {
			return site24x7.Provider(ator)
		},
	})
}
