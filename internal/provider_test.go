package internal_test

import (
	"os"
	"testing"

	"github.com/Kichiyaki/terraform-provider-woodpecker/internal"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

//nolint:unused
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"woodpecker": providerserver.NewProtocol6WithError(internal.NewProvider("test")()),
}

//nolint:unused
func testAccPreCheck(t *testing.T) {
	t.Helper()

	if v := os.Getenv("WOODPECKER_SERVER"); v == "" {
		t.Fatal("WOODPECKER_SERVER must be set for tests")
	}

	if v := os.Getenv("WOODPECKER_TOKEN"); v == "" {
		t.Fatal("WOODPECKER_TOKEN must be set for tests")
	}
}
