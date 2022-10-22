package main

import (
	"context"
	"fmt"
	"io"
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

func main() {
	// Create Docker Client Controller
	_, err := NewController()
	if err != nil {
		panic(err)
	}

	// List Containers

}

// func CreateNewContainer(image string) (string, error) {
// 	cli, err := client.NewClientWithOpts(client.FromEnv)
// 	if err != nil {
// 		fmt.Println("Unable to create docker client")
// 		panic(err)
// 	}

// 	hostBinding := nat.PortBinding{
// 		HostIP:   "0.0.0.0",
// 		HostPort: "8000",
// 	}
// 	containerPort, err := nat.NewPort("tcp", "80")
// 	if err != nil {
// 		panic("Unable to get the port")
// 	}

// 	portBinding := nat.PortMap{containerPort: []nat.PortBinding{hostBinding}}
// 	cont, err := cli.ContainerCreate(
// 		context.Background(),
// 		&container.Config{
// 			Image: image,
// 		},
// 		&container.HostConfig{
// 			PortBindings: portBinding,
// 		}, nil, "")
// 	if err != nil {
// 		panic(err)
// 	}

// 	cli.ContainerStart(context.Background(), cont.ID, types.ContainerStartOptions{})
// 	fmt.Printf("Container %s is started", cont.ID)
// 	return cont.ID, nil
// }

// func ListContainer() error {
// 	cli, err := client.NewClientWithOpts(client.FromEnv)
// 	if err != nil {
// 		panic(err)
// 	}

// 	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
// 	if err != nil {
// 		panic(err)
// 	}

// 	if len(containers) > 0 {
// 		for _, container := range containers {
// 			fmt.Printf("Container ID: %s", container.ID)
// 		}
// 	} else {
// 		fmt.Println("There are no containers running")
// 	}
// 	return nil
// }
