package e2e

import (
	"fmt"
	"os"
	"path"
	"testing"
	"time"

	lib "github.com/bazooka-ci/bazooka/commons"

	docker "github.com/fsouza/go-dockerclient"

	"github.com/bazooka-ci/bazooka/client"
	dockercmd "github.com/bywan/go-dockercommand"
)

var (
	tempDir    string
	dockerSock string
	serverHost string
)

type Bzk struct {
	Api *client.Client

	t *testing.T

	tag        string
	bzkHome    string
	dockerSock string
	scmKey     string

	dockerClient    *dockercmd.Docker
	mongoContainer  *dockercmd.Container
	serverContainer *dockercmd.Container

	repos []*Repository
}

func NewBazooka(t *testing.T) *Bzk {
	bzkHome := path.Join(tempDir, "bazooka-home")

	if err := os.MkdirAll(bzkHome, 0755); err != nil {
		t.Fatalf("Failed to allocate a temp dir as bazooka home: %v", err)
	}
	t.Logf("Created a bazooka home at %s", bzkHome)

	dockerClient, err := dockercmd.NewDocker("")
	if err != nil {
		t.Fatalf("Failed to create a docker client: %v", err)
	}

	bzk := &Bzk{
		t:            t,
		tag:          "latest",
		bzkHome:      bzkHome,
		dockerSock:   dockerSock,
		scmKey:       "",
		dockerClient: dockerClient,
	}

	bzk.startMongo()
	bzk.startServer()

	serverPort := bzk.getHostPort(bzk.serverContainer, "3000/tcp")

	timeout := 20 * time.Second
	if err := lib.WaitForTcpConnection(serverHost, serverPort, 100*time.Millisecond, timeout); err != nil {
		t.Fatalf("Couldn't connect to the bazooka API server on %s:%s after %v", serverHost, serverPort, timeout)
	}

	bzkApi, err := client.New(&client.Config{
		URL: fmt.Sprintf("http://%s:%s", serverHost, serverPort),
	})
	if err != nil {
		t.Fatalf("Failed to create a bazooka API client: %v", err)
	}
	bzk.Api = bzkApi

	return bzk
}

func (b *Bzk) Teardown() {
	b.t.Logf("Deleting the bazooka home directory: %s", b.bzkHome)
	if err := os.RemoveAll(b.bzkHome); err != nil {
		b.t.Errorf("Error while deleting bazooka home directory: %v", err)
	}

	b.t.Logf("Removing the mongo container")
	if err := b.serverContainer.Remove(&dockercmd.RemoveOptions{
		Force:         true,
		RemoveVolumes: true,
	}); err != nil {
		b.t.Errorf("Error while stopping server container: %v", err)
	}

	b.t.Logf("Removing the server container")
	if err := b.mongoContainer.Remove(&dockercmd.RemoveOptions{
		Force:         true,
		RemoveVolumes: true,
	}); err != nil {
		b.t.Errorf("Error while stopping mongo container: %v", err)
	}

	b.t.Logf("Tearing down repositories")
	for _, r := range b.repos {
		r.teardown()
	}
}

func (b *Bzk) startServer() {
	b.t.Logf("Starting a bazooka server instance")
	envMap := map[string]string{
		"BZK_HOME":       b.bzkHome,
		"BZK_DOCKERSOCK": b.dockerSock,
	}
	if len(b.scmKey) > 0 {
		envMap["BZK_SCM_KEYFILE"] = b.scmKey
	}

	container, err := b.dockerClient.Run(&dockercmd.RunOptions{
		Image:  fmt.Sprintf("bazooka/server:%s", b.tag),
		Detach: true,
		VolumeBinds: []string{
			fmt.Sprintf("%s:/bazooka", b.bzkHome),
			fmt.Sprintf("%s:/var/run/docker.sock", b.dockerSock),
		},
		Links:           []string{fmt.Sprintf("%s:mongo", b.mongoContainer.ID())},
		Env:             envMap,
		PublishAllPorts: true,
	})
	if err != nil {
		b.t.Fatalf("Failed to create the server container: %v", err)
	}
	b.t.Logf("Started a bazooka server instance")

	b.serverContainer = container
}

func (b *Bzk) startMongo() {
	b.t.Logf("Starting a mongodb instance")
	container, err := b.dockerClient.Run(&dockercmd.RunOptions{
		Image:  "mongo:3.0.2",
		Detach: true,
	})
	if err != nil {
		b.t.Fatalf("Failed to create a mongodb container: %v", err)
	}
	b.t.Logf("Started a mongodb instance")
	b.mongoContainer = container
}

func (b *Bzk) getHostPort(container *dockercmd.Container, port string) string {
	dc, err := container.Inspect()
	if err != nil {
		b.t.Fatalf("Failed to inspect a container: %v", err)
	}
	bindings, ok := dc.NetworkSettings.Ports[docker.Port(port)]
	if !ok || len(bindings) == 0 {
		b.t.Fatalf("Cannot find the host port %s was bound to", port)
	}
	return bindings[0].HostPort
}
