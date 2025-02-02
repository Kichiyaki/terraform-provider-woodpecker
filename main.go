package main

import (
	"context"
	"flag"
	"log"

	"github.com/Kichiyaki/terraform-provider-woodpecker/internal"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

//go:generate terraform fmt -recursive ./examples/
//go:generate tfplugindocs

var version = "dev"

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/Kichiyaki/woodpecker",
		Debug:   debug,
	}

	err := providerserver.Serve(context.Background(), internal.NewProvider(version), opts)
	if err != nil {
		log.Fatal(err.Error())
	}
}
