package internal_test

import (
	"testing"

	"github.com/Kichiyaki/terraform-provider-woodpecker/internal"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

//nolint:unused
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"scaffolding": providerserver.NewProtocol6WithError(internal.NewProvider("test")()),
}

//nolint:unused
func testAccPreCheck(t *testing.T) {
	t.Helper()
}
