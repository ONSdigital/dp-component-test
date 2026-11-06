package componenttest

import (
	"fmt"
	"os"

	tc "github.com/testcontainers/testcontainers-go"
)

func containerProviderFromEnv() (tc.ProviderType, error) {
	value := os.Getenv("TESTCONTAINERS_PROVIDER")
	switch value {
	case "", "default":
		return tc.ProviderDefault, nil
	case "docker":
		return tc.ProviderDocker, nil
	case "podman":
		return tc.ProviderPodman, nil
	default:
		return tc.ProviderDefault, fmt.Errorf("invalid TESTCONTAINERS_PROVIDER %q: expected default, docker, or podman", value)
	}
}
