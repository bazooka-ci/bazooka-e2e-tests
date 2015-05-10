package e2e

import (
	"fmt"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	tempDir = os.Getenv("BZK_E2E_TEMP")
	if len(tempDir) == 0 {
		fmt.Printf("$BZK_E2E_TEMP must be set to the location which will be used by the tests as a temporary bazooka home\n")
		os.Exit(-1)
	}

	dockerSock = os.Getenv("BZK_E2E_DOCKER_SOCK")
	if len(dockerSock) == 0 {
		dockerSock = "/var/run/docker.sock"
		fmt.Printf("$BZK_E2E_DOCKER_SOCK is not set, defaulting to %s\n", dockerSock)
	}

	serverHost = os.Getenv("BZK_E2E_HOST")
	if len(serverHost) == 0 {
		fmt.Printf("$BZK_E2E_HOST must be set to the host name or ip address of the machine running the tests\n")
		os.Exit(-1)
	}

	os.Exit(m.Run())
}
