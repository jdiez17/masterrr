package main

import (
	"log"
	dockerc "github.com/samalba/dockerclient"
	"runtime/debug"
)

func startContainer(image string) (string, HTTPError) {
	id, err := State.Docker.CreateContainer(&dockerc.ContainerConfig{
		Image: image,
		NetworkDisabled: false,
	}, "", nil)
	if err != nil {
		log.Println("ERR: CreateContainer:", err)
		return "", NewHTTPError("Unable to create container.", 500)
	}

	err = State.Docker.StartContainer(id, nil)
	if err != nil {
		log.Println("ERR: StartContainer:", err)
		return "", NewHTTPError("Unable to start container.", 500)
	}
	log.Println("INFO: Container", id[:16], "started.")

	return id, nil
}

func stopContainer(id string) HTTPError {
	err := State.Docker.KillContainer(id, "SIGKILL")
	if err != nil {
		log.Println("WARN: KillContainer:", err)
		return NewHTTPError("Unable to kill container.", 500)
	}
	log.Println("INFO: Container", id[:16], "killed.")

	err = State.Docker.RemoveContainer(id, true, false)
	if err != nil {
		log.Println("WARN: RemoveContainer:", err)
		return NewHTTPError("Unable to remove container.", 500)
	}
	log.Println("INFO: Container", id[:16], "removed.")

	return nil
}

func getContainerIP(id string) (string, HTTPError) {
	info, err := State.Docker.InspectContainer(id)
	if err != nil {
		log.Println("ERR: InspectContainer:", err)
		return "", NewHTTPError("Unable to inspect container.", 500)
	}
	ip := info.NetworkSettings.IPAddress
	log.Println("INFO: IP address:", ip)

	return ip, nil
}
