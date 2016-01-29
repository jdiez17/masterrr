package main

import (
	"net/http"
	"log"
	"strings"
	dockerc "github.com/samalba/dockerclient"
)

const servicePort = "8081"
const port = "8082"
const image = "jdiez/dockerrr"
var docker *dockerc.DockerClient

func startHandler(r *http.Request) (interface{}, HTTPError) {
	// Spin up a container for this request
	id, err := docker.CreateContainer(&dockerc.ContainerConfig{
		Image: image,
		NetworkDisabled: false,
	}, "test", nil)
	if err != nil {
		log.Println("ERR: CreateContainer:", err)
		return nil, NewHTTPError("Unable to create container.", 500)
	}

	err = docker.StartContainer(id, nil)
	if err != nil {
		log.Println("ERR: StartContainer:", err)
		return nil, NewHTTPError("Unable to start container.", 500)
	}
	log.Println("INFO: Container", id, "started.")

	info, err := docker.InspectContainer(id)
	if err != nil {
		log.Println("ERR: InspectContainer:", err)
		return nil, NewHTTPError("Unable to inspect container.", 500)
	}
	ip := info.NetworkSettings.IPAddress
	log.Println("INFO: IP address:", ip)

	return StartResponse{
		Endpoint: ip + ":" + servicePort,
		ContainerID: id,
	}, nil
}

func stopHandler(r *http.Request) (interface{}, HTTPError) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 3 {
		return nil, NewHTTPError("ID argument not given.", 400)
	}

	id := parts[2]
	err := docker.KillContainer(id, "SIGKILL")
	if err != nil {
		log.Println("WARN: KillContainer:", err)
		return nil, NewHTTPError("Unable to kill container.", 500)
	}
	log.Println("INFO: Container", id, "killed.")

	err = docker.RemoveContainer(id, true, false)
	if err != nil {
		log.Println("WARN: RemoveContainer:", err)
		return nil, NewHTTPError("Unable to remove container.", 500)
	}
	log.Println("INFO: Container", id, "removed.")

	return GenericResponse{Success: true}, nil
}

func main() {
	var err error

	// Connect to the Docker daemon
	docker, err = dockerc.NewDockerClient("unix:////var/run/docker.sock", nil)
	if err != nil {
		log.Fatal(err)
	}

	// Ensure we're actually connected
	info, err := docker.Info()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Running at", info.Name, "on", info.OperatingSystem)

	http.HandleFunc("/start/", jsonM(startHandler))
	http.HandleFunc("/stop/", jsonM(stopHandler))
	log.Println("masterrr: listening on", ":" + port)
	http.ListenAndServe(":" + port, nil)
}
