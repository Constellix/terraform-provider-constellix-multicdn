package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/constellix/terraform-provider-constellix-multicdn/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
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

	fmt.Printf("Version: %s, Commit: %s\n", version, commit)

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/constellix/constellix-multicdn",
		Debug:   debug,
	}

	err := providerserver.Serve(context.Background(), provider.New, opts)

	if err != nil {
		log.Fatal(err.Error())
	}
}
