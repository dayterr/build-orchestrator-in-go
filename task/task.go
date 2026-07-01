package task

import (
	"context"
	"io"
	"log"
	"math"
	"os"

	"github.com/docker/docker/pkg/stdcopy"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
)

func (d *Docker) Run() DockerResult {
	ctx := context.Background()
	po := image.PullOptions{}
	reader, err := d.Client.ImagePull(ctx, d.Config.Image, po)
	if err != nil {
		log.Printf("Failed to pull image %s: %v\n", d.Config.Image, err)
		return DockerResult{Error: err}
	}

	_, err = io.Copy(os.Stdout, reader)
	if err != nil {
		log.Printf("Failed to read image %s: %v\n", d.Config.Image, err)
		return DockerResult{Error: err}
	}

	rp := container.RestartPolicy{
		Name: d.Config.RestartPolicy,
	}

	r := container.Resources{
		Memory:   d.Config.Memory,
		NanoCPUs: int64(d.Config.CPU * math.Pow(10, 9)),
	}

	cc := container.Config{
		Image:        d.Config.Image,
		Tty:          false,
		Env:          d.Config.Env,
		ExposedPorts: d.Config.ExposedPorts,
	}

	hc := container.HostConfig{
		RestartPolicy:   rp,
		Resources:       r,
		PublishAllPorts: true,
	}

	resp, err := d.Client.ContainerCreate(ctx, &cc, &hc, nil, nil, d.Config.Name)
	if err != nil {
		log.Printf("Failed to create container unsing image %s: %v\n", d.Config.Image, err)
		return DockerResult{Error: err}
	}

	so := container.StartOptions{}
	err = d.Client.ContainerStart(ctx, resp.ID, so)
	if err != nil {
		log.Printf("Failed to start container %s: %v\n", resp.ID, err)
		return DockerResult{Error: err}
	}

	lo := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
	}
	out, err := d.Client.ContainerLogs(ctx, resp.ID, lo)
	if err != nil {
		log.Printf("Failed to get logs for container %s: %v\n", resp.ID, err)
	}

	_, err = stdcopy.StdCopy(os.Stdout, os.Stderr, out)
	if err != nil {
		log.Printf("Failed to copy output from container: %v\n", err)
		return DockerResult{Error: err}
	}

	return DockerResult{ContainerID: resp.ID, Action: "start", Result: "success"}
}

func (d *Docker) Stop(id string) DockerResult {
	log.Printf("Stopping container %s\n", id)

	ctx := context.Background()
	err := d.Client.ContainerStop(ctx, id, container.StopOptions{})
	if err != nil {
		log.Printf("Failed to stop container %s: %v\n", id, err)
		return DockerResult{Error: err}
	}

	ro := container.RemoveOptions{
		RemoveVolumes: true,
		RemoveLinks:   false,
		Force:         true,
	}
	err = d.Client.ContainerRemove(ctx, id, ro)
	if err != nil {
		log.Printf("Failed to remove container %s: %v\n", id, err)
		return DockerResult{Error: err}
	}

	return DockerResult{Action: "stop", Result: "success"}
}
