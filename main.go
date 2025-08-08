package main

import (
	"context"
	"flag"
	"log"

	"github.com/constellix/terraform-provider-constellix-multicdn/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	// these will be set by the goreleaser configuration
	// to appropriate values for the compiled binary.
	version string = "dev"
	commit  string = "none"
)

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	tflog.Info(context.Background(), "Provider version info", map[string]interface{}{
		"version": version,
		"commit":  commit,
	})

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/constellix/constellix-multicdn",
		Debug:   debug,
	}

	err := providerserver.Serve(context.Background(), provider.New, opts)

	if err != nil {
		log.Fatal(err.Error())
	}
}
