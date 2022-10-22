package main

import (
	"fmt"
	"testing"
)

func TestEnsureImage(t *testing.T) {
	c, err := NewController()

	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	err = c.EnsureImage("python:3.10-slim-buster")

	if err != nil {
		t.Error(err)
	}
}

func TestContainerRun(t *testing.T) {
	c, err := NewController()

	if err != nil {
		t.Error(err)
	}

	statusCode, body, err := c.ContainerRunAndClean("alpine", []string{"echo", "hello world"}, []VolumeMount{})

	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if body != "hello world\r\n" {
		t.Errorf("Expected 'hello world'; received %q\n", body)
	}

	if statusCode != 0 {
		t.Errorf("Expect status to be 0; received %q\n", statusCode)
	}
}

func TestContainerRunWithMountedVolume(t *testing.T) {
	c, err := NewController()

	if err != nil {
		t.Error(err)
	}

	// Create mounted volume
	mounts := []VolumeMount{
		{
			HostPath:   "/home/korbersa/Downloads",
			TargetPath: "/test",
		},
	}

	statusCode, body, err := c.ContainerRunAndClean("python:3.10-slim-buster", []string{"ls", "-lah", "/test"}, mounts)

	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	fmt.Println(body)
	if body == "" {
		t.Errorf("Expected files; received %q\n", body)
	}

	if statusCode != 0 {
		t.Errorf("Expect status to be 0; received %q\n", statusCode)
	}
}
