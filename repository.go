package e2e

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"testing"

	dockercmd "github.com/bywan/go-dockercommand"
)

var (
	repoIndex = 0
)

type Repository struct {
	index int
	t     *testing.T

	location string

	dockerClient *dockercmd.Docker
	container    *dockercmd.Container
	port         string
}

func (b *Bzk) NewRepository() *Repository {
	index := repoIndex
	repoIndex++

	location := path.Join(tempDir, fmt.Sprintf("bazooka-repo-%d", index))

	os.RemoveAll(location)
	if err := os.MkdirAll(location, 0755); err != nil {
		b.t.Fatalf("Failed to allocate a temp dir for repository %d: %v", index, err)
	}
	b.t.Logf("Created a repository %d home at %s", index, location)

	repo := &Repository{
		t:            b.t,
		index:        index,
		location:     location,
		dockerClient: b.dockerClient,
	}

	repo.cmd("git", "init")

	b.t.Logf("Starting a git server instance for repository %d", index)
	container, err := b.dockerClient.Run(&dockercmd.RunOptions{
		Image:  "bazooka/e2e-git",
		Detach: true,
		VolumeBinds: []string{
			fmt.Sprintf("%s:/repo", location),
		},
		PublishAllPorts: true,
	})
	if err != nil {
		b.t.Fatalf("Failed to create the git server container for repository %d: %v", index, err)
	}
	b.t.Logf("Started a git server instance for repository %d, id: %s", index, container.ID())
	repo.ContainerLog("<git-srv>", container)

	port := b.getHostPort(container, "9418/tcp")

	repo.container = container
	repo.port = port

	b.repos = append(b.repos, repo)
	return repo
}

func (r *Repository) teardown() {
	r.t.Logf("Deleting the repository directory: %s", r.location)
	if err := os.RemoveAll(r.location); err != nil {
		r.t.Errorf("Error while deleting the repository directory: %v", err)
	}

	r.t.Logf("Removing the git server container")
	if err := r.container.Remove(&dockercmd.RemoveOptions{
		Force:         true,
		RemoveVolumes: true,
	}); err != nil {
		r.t.Errorf("Error while removing the git server container: %v", err)
	}
}

func (r *Repository) CloneURL() string {
	return fmt.Sprintf("git://%s:%s/", "boot2docker", r.port)
}

func (r *Repository) ImportFile(src, dst string) {
	if err := copyFileContents(src, filepath.Join(r.location, dst)); err != nil {
		r.t.Fatalf("Error while copying file %s to the repository %d: %v", src, r.index, err)
	}
}

func (r *Repository) ImportDir(src string) {
	if err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		switch {
		case path == src:
			return nil
		case info.IsDir():
			dst := filepath.Join(r.location, strings.TrimPrefix(path, src))
			if err := os.MkdirAll(dst, 0755); err != nil {
				return err
			}
		default:
			r.ImportFile(path, strings.TrimPrefix(path, src))
		}

		return nil
	}); err != nil {
		r.t.Fatalf("Error while importing dir %s: %v", src, err)
	}
}

func (r *Repository) cmd(cmd ...string) {
	r.t.Logf("Executing command %v", cmd)
	container, err := r.dockerClient.Run(&dockercmd.RunOptions{
		Image: "bazooka/e2e-git",
		VolumeBinds: []string{
			fmt.Sprintf("%s:/repo", r.location),
		},
		Cmd: cmd,
	})

	if err != nil {
		r.t.Fatalf("Failed to execute the command %v: %v", cmd, err)
	}
	defer func() {
		r.t.Logf("Removing command container %s", container.ID())
		if err := container.Remove(&dockercmd.RemoveOptions{
			Force:         true,
			RemoveVolumes: true,
		}); err != nil {
			r.t.Errorf("Error while removing the command container %s: %v", container.ID(), err)
		}
	}()
	r.ContainerLog("<cmd>", container)

	exitCode, err := container.Wait()
	if err != nil {
		r.t.Fatalf("Failed to retrieve the exit code of command %v: %v", cmd, err)
	}
	if exitCode != 0 {
		r.t.Fatalf("Failed to execute the command %v: exit code %d", cmd, exitCode)
	}
}

func (r *Repository) ContainerLog(prefix string, container *dockercmd.Container) {
	reader, writer := io.Pipe()
	container.StreamLogs(writer)
	go func(reader io.Reader) {
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			r.t.Logf("[%s] %s \n", prefix, scanner.Text())

		}
		if err := scanner.Err(); err != nil {
			r.t.Errorf("There was an error with the scanner in attached container: %v", err)
		}
	}(reader)
}

// copyFileContents copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file.
func copyFileContents(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return
	}
	err = out.Sync()

	// This is an ugly hack to fix sporadic problems with boot2docker and virtualbox
	// where the copied file is sometimes empty
	exec.Command("sync").Run()

	return
}
