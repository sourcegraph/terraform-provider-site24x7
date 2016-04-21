package main

import (
	"github.com/hashicorp/terraform/plugin"
	"github.com/sourcegraph/terraform-provider-site24x7/site24x7"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: site24x7.Provider,
	})
}
