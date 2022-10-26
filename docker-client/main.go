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
	"github.com/docker/docker/pkg/stdcopy"
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
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	reader, err := c.cli.ContainerLogs(context.Background(), id, types.ContainerLogsOptions{
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

func (c *Controller) ContainerRunDetached(image string, volumes []VolumeMount) (id string, err error) {
	// Start the container
	hostConfig := container.HostConfig{}

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
		Image: image,
	}, &hostConfig, nil, nil, "")

	if err != nil {
		return "", err
	}

	err = c.cli.ContainerStart(context.Background(), resp.ID, types.ContainerStartOptions{})
	if err != nil {
		return resp.ID, err
	}

	// List containers
	err = c.ListContainers()
	if err != nil {
		return resp.ID, err
	}

	return resp.ID, nil
}

func (c *Controller) Exec(containerID string, command []string) error {
	// Create exec config
	fmt.Println("Create config")
	config := types.ExecConfig{
		AttachStderr: true,
		AttachStdout: true,
		Cmd:          command,
	}

	// Create the exec
	fmt.Println("Create the exec")
	responseID, err := c.cli.ContainerExecCreate(context.Background(), containerID, config)
	if err != nil {
		return err
	}

	// Start the exec
	fmt.Println("Start the exec")
	err = c.cli.ContainerExecStart(context.Background(), responseID.ID, types.ExecStartCheck{})
	if err != nil {
		return err
	}

	// Get status
	time.Sleep(5)
	inspectResult, err := c.cli.ContainerExecInspect(context.Background(), responseID.ID)
	if err != nil {
		return err
	}
	fmt.Println("Running:", inspectResult.Running)
	fmt.Println("ExitCode:", inspectResult.ExitCode)
	fmt.Println("Pid:", inspectResult.Pid)

	// Get logs
	log, err := c.ContainerLog(containerID)
	if err != nil {
		return err
	}
	fmt.Println("Logs:", log)

	// Get the exec response
	fmt.Println("Get the response")
	var results ExecResult
	results, err = c.InspectExecResp(containerID)

	fmt.Println("Stdout:", results.StdOut)
	fmt.Println("Stderr:", results.StdErr)
	fmt.Println("Exit code:", results.ExitCode)

	return err
}

func (c *Controller) InspectExecResp(containerID string) (ExecResult, error) {
	var execResult ExecResult

	fmt.Println("Exec attach")
	resp, err := c.cli.ContainerExecAttach(context.Background(), containerID, types.ExecStartCheck{})
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

	case <-context.Background().Done():
		return execResult, context.Background().Err()
	}

	fmt.Println("Read stdout")
	stdout, err := ioutil.ReadAll(&outBuf)
	if err != nil {
		return execResult, err
	}
	stderr, err := ioutil.ReadAll(&errBuf)
	if err != nil {
		return execResult, err
	}

	res, err := c.cli.ContainerExecInspect(context.Background(), containerID)
	if err != nil {
		return execResult, err
	}

	execResult.ExitCode = res.ExitCode
	execResult.StdOut = string(stdout)
	execResult.StdErr = string(stderr)
	return execResult, nil
}

func (c *Controller) StopContainer(containerID string) (string, error) {
	// Get the log
	body, _ := c.ContainerLog(containerID)

	err := c.cli.ContainerRemove(context.Background(), containerID, types.ContainerRemoveOptions{})

	if err != nil {
		fmt.Printf("Unable to remove container %q: %q\n", containerID, err)
	}

	return body, err
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
