package main

import (
	"time"
	"log"
	"strings"
	"net/http"
)

const interval = 10 * time.Second

func testLiveness(id string, ip string) {
	url := "http://" + ip + ":" + servicePort + "/control/ping"

	// Inspect container to make sure it's running
	info, err := State.Docker.InspectContainer(id)
	if err != nil {
		// It's kill, remove it from our state
		log.Println("WARN: Removing dead container", id[:16])
		State.RemoveContainer(id, ip)
		return
	}
	if !info.State.Running {
		// Don't want to keep dead containers around
		log.Println("WARN: Removing non-running container", id[:16])
		stopContainer(id)
		State.RemoveContainer(id, ip)
		return
	}

	// Allow the service to boot up - don't run liveness check for a minute
	if time.Since(info.State.StartedAt) < time.Minute {
		log.Println("INFO: Liveness: Skipping container", id[:16], "because it was launched recently.")
		return
	}

	res, err := http.Get(url)
	if res != nil {
		if res.StatusCode != 200 {
			log.Println("WARN: Liveness: check for container", id[:16], "returned code:", res.StatusCode, "(err", err, ")")

			stopContainer(id)
			State.RemoveContainer(id, ip)
			return
		}
	}
	if err != nil {
		log.Println("WARN: Liveness: check for", id[:16], "failed with:", err)
		stopContainer(id)
		State.RemoveContainer(id, ip)
	}

	log.Println("INFO: Liveness: check for container", id[:16], " OK")
}

func monitor() {
	// Every `interval`, send a request to all the containers we're monitoring.
	for range time.Tick(interval) {
		containers, err := State.MonitoredContainers()
		if err != nil {
			continue
		}

		for _, container := range containers {
			parts := strings.Split(container, ":")
			if len(parts) != 2 {
				log.Println("WARN: Error parsing container entry:", container)
			}

			go testLiveness(parts[0], parts[1])
		}
	}
}
