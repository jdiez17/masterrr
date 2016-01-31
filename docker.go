package main

import (
	dockerc "github.com/samalba/dockerclient"
	"log"
)

var docker *dockerc.DockerClient = nil

type containerInfo struct {
	ID, IP string
}

func dockerConnect() {
	var err error

	// Connect to the Docker daemon
	docker, err = dockerc.NewDockerClient("unix:////var/run/docker.sock", nil)
	if err != nil {
		log.Fatal("ERR: Docker says: ", err)
	}

	// Ensure we're actually connected
	info, err := docker.Info()
	if err != nil {
		log.Fatal("ERR: Docker says: ", err)
	}
	log.Println("Running at", info.Name, "on", info.OperatingSystem)
}

func getContainerIP(id string) (string, HTTPError) {
	info, err := docker.InspectContainer(id)
	if err != nil {
		log.Println("ERR: InspectContainer:", err)
		return "", NewHTTPError("Unable to inspect container.", 500)
	}
	ip := info.NetworkSettings.IPAddress
	log.Println("INFO: IP address:", ip)

	return ip, nil
}

func startContainer(image string) (*containerInfo, HTTPError) {
	id, err := docker.CreateContainer(&dockerc.ContainerConfig{
		Image:           image,
		NetworkDisabled: false,
		Labels: map[string]string{"role": "worker"},
	}, "", nil)
	if err != nil {
		log.Println("ERR: CreateContainer:", err)
		return nil, NewHTTPError("Unable to create container.", 500)
	}

	err = docker.StartContainer(id, &dockerc.HostConfig{
		Memory: 100 * 1024 * 1024,
		KernelMemory: 1 * 1024,
		CpuQuota: 10000,
	})
	if err != nil {
		log.Println("ERR: StartContainer:", err)
		return nil, NewHTTPError("Unable to start container.", 500)
	}
	log.Println("INFO: Container", id[:16], "started.")

	ip, err2 := getContainerIP(id)
	if err2 != nil {
		stopContainer(id)
		return nil, err2
	}

	// Add to the list of containers managed by this process
	if err := State.AddContainer(id); err != nil {
		stopContainer(id)
		return nil, NewHTTPError("Unable to add container to Redis.", 500)
	}

	return &containerInfo{ID: id, IP: ip}, nil
}

func stopContainer(id string) HTTPError {
	State.RemoveContainer(id)

	err := docker.KillContainer(id, "SIGKILL")
	if err != nil {
		log.Println("WARN: KillContainer:", err)
		return NewHTTPError("Unable to kill container.", 500)
	}
	log.Println("INFO: Container", id[:16], "killed.")

	err = docker.RemoveContainer(id, true, false)
	if err != nil {
		log.Println("WARN: RemoveContainer:", err)
		return NewHTTPError("Unable to remove container.", 500)
	}
	log.Println("INFO: Container", id[:16], "removed.")

	return nil
}
