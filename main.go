package main

import (
	"github.com/eliaszs/terraform-provider-firebase/firebase"
	"github.com/hashicorp/terraform/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: firebase.Provider})
}
