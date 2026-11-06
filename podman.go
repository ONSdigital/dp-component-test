package componenttest

// Register Podman as an optional Testcontainers runtime provider.
// See https://golang.testcontainers.org/system_requirements/using_podman/#using-the-podman-module
import _ "github.com/testcontainers/testcontainers-go/modules/podman" // Required to register the provider.
