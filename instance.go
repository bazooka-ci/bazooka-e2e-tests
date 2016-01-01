package e2e

import (
	"fmt"
	"os"
	"path"
	"testing"
	"time"

	lib "github.com/bazooka-ci/bazooka/commons"

	docker "github.com/fsouza/go-dockerclient"

	"io"

	"github.com/bazooka-ci/bazooka/client"
	dockercmd "github.com/bywan/go-dockercommand"
)

const (
	NET_NAME = "bzk_e2e_net"
)

var (
	tempDir    string
	dockerSock string
	serverHost string
	apiPort    string
	syslogPort string
)

type Bzk struct {
	Api *client.Client

	t *testing.T

	ts int64

	tag        string
	bzkHome    string
	dockerSock string
	scmKey     string

	dockerClient        *dockercmd.Docker
	mongoContainer      *dockercmd.Container
	serverContainer     *dockercmd.Container
	mongoContainerName  string
	serverContainerName string

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
		ts:           time.Now().UnixNano(),
	}

	bzk.createNetwork()

	bzk.startMongo()
	// time.Sleep(5 * time.Second)
	bzk.startServer()

	timeout := 20 * time.Second
	if err := lib.WaitForTcpConnection(fmt.Sprintf("%s:%s", serverHost, apiPort), 100*time.Millisecond, timeout); err != nil {
		t.Fatalf("Couldn't connect to the bazooka API server on %s:%s after %v", serverHost, apiPort, timeout)
	}

	{
		r, w := io.Pipe()
		bzk.serverContainer.StreamLogs(w)
		scanner := lib.NewScanner(r)
		go func() {
			for scanner.Scan() {
				message := scanner.Text()
				t.Log(message)
			}
		}()
	}
	bzkApi, err := client.New(&client.Config{
		URL: fmt.Sprintf("http://%s:%s", serverHost, apiPort),
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

func (b *Bzk) createNetwork() {
	ns, err := b.dockerClient.Networks()
	if err != nil {
		b.t.Fatalf("Failed to list docker networks: %v", err)
	}

	for _, n := range ns {
		if n.Name == NET_NAME {
			b.t.Logf("Found existing %s network", NET_NAME)
			return
		}
	}
	if _, err = b.dockerClient.CreateNetwork(docker.CreateNetworkOptions{
		Name:   NET_NAME,
		Driver: "bridge",
	}); err != nil {
		b.t.Fatalf("Error while creating %s bridge network: %v", NET_NAME, err)
	}
	b.t.Logf("Created bridge network %s", NET_NAME)
}

func (b *Bzk) startServer() {
	b.t.Logf("Starting a bazooka server instance")

	b.serverContainerName = fmt.Sprintf("bzk_server_e2e_%d", b.ts)

	envMap := map[string]string{
		"BZK_HOME":       b.bzkHome,
		"BZK_DOCKERSOCK": b.dockerSock,
		"BZK_NETWORK":    NET_NAME,
		"BZK_SYSLOG_URL": fmt.Sprintf("tcp://%s:%s", serverHost, syslogPort),
		"BZK_API_URL":    fmt.Sprintf("http://%s:3000", b.serverContainerName),
		"BZK_DB_URL":     fmt.Sprintf("%s:27017", b.mongoContainerName),
	}
	if len(b.scmKey) > 0 {
		envMap["BZK_SCM_KEYFILE"] = b.scmKey
	}

	b.t.Logf("mongo id=%s\n", b.mongoContainer.ID())

	container, err := b.dockerClient.Run(&dockercmd.RunOptions{
		Name:   b.serverContainerName,
		Image:  fmt.Sprintf("bazooka/server:%s", b.tag),
		Detach: true,
		VolumeBinds: []string{
			fmt.Sprintf("%s:/bazooka", b.bzkHome),
			fmt.Sprintf("%s:/var/run/docker.sock", b.dockerSock),
		},
		Env:         envMap,
		NetworkMode: NET_NAME,
		PortBindings: map[docker.Port][]docker.PortBinding{
			"3000/tcp": {{HostPort: apiPort}},
			"3001/tcp": {{HostPort: syslogPort}},
		},
	})
	if err != nil {
		b.t.Fatalf("Failed to create the server container: %v", err)
	}
	b.t.Logf("Started a bazooka server instance")

	b.serverContainer = container
}

func (b *Bzk) startMongo() {
	b.t.Logf("Starting a mongodb instance")

	b.mongoContainerName = fmt.Sprintf("bzk_db_e2e_%d", b.ts)

	container, err := b.dockerClient.Run(&dockercmd.RunOptions{
		Name:        b.mongoContainerName,
		Image:       "mongo:3.0.2",
		Detach:      true,
		Cmd:         []string{"--smallfiles"},
		NetworkMode: NET_NAME,
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
