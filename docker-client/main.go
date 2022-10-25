package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
)

type Controller struct {
	cli *client.Client
}

type VolumeMount struct {
	HostPath   string
	TargetPath string
}

type ExecResult struct {
	StdOut   string
	StdErr   string
	ExitCode int
}

func NewController() (c *Controller, err error) {
	c = new(Controller)

	c.cli, err = client.NewClientWithOpts(client.FromEnv)

	if err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Controller) EnsureImage(image string) (err error) {
	reader, err := c.cli.ImagePull(context.Background(), image, types.ImagePullOptions{})

	if err != nil {
		return err
	}
	defer reader.Close()
	io.Copy(os.Stdout, reader)
	return nil
}

func (c *Controller) ContainerLog(id string) (result string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	reader, err := c.cli.ContainerLogs(ctx, id, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true})

	if err != nil {
		return "", err
	}

	buffer, err := io.ReadAll(reader)

	if err != nil && err != io.EOF {
		return "", err
	}

	return string(buffer), nil
}

func (c *Controller) ContainerRun(image string, command []string, volumes []VolumeMount) (id string, err error) {
	hostConfig := container.HostConfig{}

	//	hostConfig.Mounts = make([]mount.Mount,0);

	var mounts []mount.Mount

	for _, volume := range volumes {
		mount := mount.Mount{
			Type:   mount.TypeBind,
			Source: volume.HostPath,
			Target: volume.TargetPath,
		}
		mounts = append(mounts, mount)
	}

	hostConfig.Mounts = mounts

	resp, err := c.cli.ContainerCreate(context.Background(), &container.Config{
		Tty:   true,
		Image: image,
		Cmd:   command,
	}, &hostConfig, nil, nil, "")

	if err != nil {
		return "", err
	}

	err = c.cli.ContainerStart(context.Background(), resp.ID, types.ContainerStartOptions{})
	if err != nil {
		return "", err
	}

	return resp.ID, nil
}

func (c *Controller) ContainerWait(id string) (state int64, err error) {
	resultC, errC := c.cli.ContainerWait(context.Background(), id, "")
	select {
	case err := <-errC:
		return 0, err
	case result := <-resultC:
		return result.StatusCode, nil
	}
}

func (c *Controller) ListContainers() error {
	// List containers
	containers, err := c.cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}

	if len(containers) > 0 {
		for _, container := range containers {
			fmt.Printf("Container ID: %s\n", container.ID)
		}
	} else {
		fmt.Println("There are no containers running")
	}
	return nil
}

func (c *Controller) ContainerRunAndClean(image string, command []string, volumes []VolumeMount) (statusCode int64, body string, err error) {
	// Start the container
	id, err := c.ContainerRun(image, command, volumes)
	if err != nil {
		return statusCode, body, err
	}

	// List containers
	err = c.ListContainers()
	if err != nil {
		return statusCode, body, err
	}

	// Wait for it to finish
	statusCode, err = c.ContainerWait(id)
	if err != nil {
		return statusCode, body, err
	}

	// Get the log
	body, _ = c.ContainerLog(id)

	err = c.cli.ContainerRemove(context.Background(), id, types.ContainerRemoveOptions{})

	if err != nil {
		fmt.Printf("Unable to remove container %q: %q\n", id, err)
	}

	return statusCode, body, err
}

func (c *Controller) Exec(containerID string, command []string) (string, error) {
	// Create exec config
	config := types.ExecConfig{
		AttachStderr: true,
		AttachStdout: true,
		Cmd:          command,
	}

	// Create the exec
	execID, err := c.cli.ContainerExecCreate(context.Background(), containerID, config)

	// Start the exec
	c.cli.ContainerExecStart(context.Background())
}

func InspectExecResp(ctx context.Context, id string) (ExecResult, error) {
	var execResult ExecResult
	docker, err := client.NewEnvClient()
	if err != nil {
		return execResult, err
	}
	defer closer(docker)

	resp, err := docker.ContainerExecAttach(ctx, id, types.ExecConfig{})
	if err != nil {
		return execResult, err
	}
	defer resp.Close()

	// read the output
	var outBuf, errBuf bytes.Buffer
	outputDone := make(chan error)

	go func() {
		// StdCopy demultiplexes the stream into two buffers
		_, err = stdcopy.StdCopy(&outBuf, &errBuf, resp.Reader)
		outputDone <- err
	}()

	select {
	case err := <-outputDone:
		if err != nil {
			return execResult, err
		}
		break

	case <-ctx.Done():
		return execResult, ctx.Err()
	}

	stdout, err := ioutil.ReadAll(&outBuf)
	if err != nil {
		return execResult, err
	}
	stderr, err := ioutil.ReadAll(&errBuf)
	if err != nil {
		return execResult, err
	}

	res, err := docker.ContainerExecInspect(ctx, id)
	if err != nil {
		return execResult, err
	}

	execResult.ExitCode = res.ExitCode
	execResult.StdOut = string(stdout)
	execResult.StdErr = string(stderr)
	return execResult, nil
}

func main() {
	// Create Docker Client Controller
	_, err := NewController()
	if err != nil {
		panic(err)
	}

	// List Containers

}

// for exec: https://stackoverflow.com/questions/52774830/docker-exec-command-from-golang-api
// examples: https://willschenk.com/articles/2021/controlling_docker_in_golang/
// sdk: https://pkg.go.dev/github.com/docker/docker@v20.10.20+incompatible/client
