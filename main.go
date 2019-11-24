package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	//"github.com/hashicorp/terraform/plugin"
	//"github.com/hashicorp/terraform/terraform"
	//"github.com/sourcegraph/terraform-provider-site24x7/site24x7"
	"github.com/sourcegraph/terraform-provider-site24x7/site24x7/oauth"
)

var oauthFile = flag.String("oauth-file", "", "(required) path to the oauth-file, will be created if it doesn't exist")
var initOauth = flag.Bool("init-oauth", false, "if set initializes oauth-file and exits")

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

	if *initOauth {
		os.Exit(0)
	}

	refreshHandle := func(w http.ResponseWriter, req *http.Request) {
		err := ator.Refresh()
		if err != nil {
			fmt.Fprintf(w, "error refreshing access token: %v\n", err)
		} else {
			fmt.Fprintf(w, "access token refreshed\n",)
		}
	}

	//plugin.Serve(&plugin.ServeOpts{
	//	ProviderFunc: func() terraform.ResourceProvider {
	//		return site24x7.Provider(ator)
	//	},
	//})

	http.HandleFunc("/refresh", refreshHandle)

	log.Fatal(http.ListenAndServe(":8484", nil))
}
