package componenttest

import (
	"fmt"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	tc "github.com/testcontainers/testcontainers-go"
)

func TestContainerProviderFromEnv(t *testing.T) {
	tests := []struct {
		name  string
		value string
		unset bool
		want  tc.ProviderType
	}{
		{name: "unset", unset: true, want: tc.ProviderDefault},
		{name: "empty", want: tc.ProviderDefault},
		{name: "default", value: "default", want: tc.ProviderDefault},
		{name: "docker", value: "docker", want: tc.ProviderDocker},
		{name: "podman", value: "podman", want: tc.ProviderPodman},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Convey("Given TESTCONTAINERS_PROVIDER is "+tt.name, t, func() {
				t.Setenv("TESTCONTAINERS_PROVIDER", tt.value)
				if tt.unset {
					So(os.Unsetenv("TESTCONTAINERS_PROVIDER"), ShouldBeNil)
				}

				Convey("When the container provider is selected", func() {
					provider, err := containerProviderFromEnv()

					Convey("Then the expected provider is returned without an error", func() {
						So(err, ShouldBeNil)
						So(provider, ShouldEqual, tt.want)
					})
				})
			})
		})
	}
}

func TestContainerProviderFromEnvInvalid(t *testing.T) {
	for _, value := range []string{"invalid", "Podman", " docker "} {
		t.Run(value, func(t *testing.T) {
			Convey("Given TESTCONTAINERS_PROVIDER has an invalid value", t, func() {
				t.Setenv("TESTCONTAINERS_PROVIDER", value)

				Convey("When the container provider is selected", func() {
					_, err := containerProviderFromEnv()

					Convey("Then a configuration error identifies the value and supported providers", func() {
						So(err, ShouldBeError, fmt.Sprintf("invalid TESTCONTAINERS_PROVIDER %q: expected default, docker, or podman", value))
					})
				})
			})
		})
	}
}
