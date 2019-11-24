package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/sourcegraph/terraform-provider-site24x7/site24x7/oauth"
)

func main() {
	var oauthFile string

	if len(os.Args) != 2 {
		homeOir, err := os.UserHomeDir()
		if err != nil {
			log.Fatal(err)
		}
		oauthFile = filepath.Join(homeOir, ".site24x7-oauth.json")
	} else {
		oauthFile = os.Args[1]
	}

	_, err := oauth.NewAuthenticator(oauthFile)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("set env var SITE24X7_AUTHTOKEN_FILE=%s", oauthFile)
}
