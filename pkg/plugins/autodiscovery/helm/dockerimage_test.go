package helm

import (
	"testing"
)

func TestGetRegistryEndpoint(t *testing.T) {

	tests := []struct {
		description      string
		image            string
		expectedRegistry string
	}{
		{
			image:            "updatecli/updatecli",
			expectedRegistry: "index.docker.io",
		},
		{
			image:            "ghcr.io/updatecli/updatecli",
			expectedRegistry: "ghcr.io",
		},
		{
			image:            "myregistry.io/nested/namespace/image-name",
			expectedRegistry: "myregistry.io",
		},
	}

	for i := range tests {
		got := sanitizeRegistryEndpoint(tests[i].image)

		if got != tests[i].expectedRegistry {
			t.Errorf("Expected %q but got %q", tests[i].expectedRegistry, got)
		}
	}
}
