package main

import (
	"log"
	"net/http"
	"sync"
	"time"
)

// Prime numbers so they don't coincide
const interval = 7 * time.Second
const timeout = 13 * time.Second

var pending map[string]bool
var mutex sync.Mutex

func setPending(id string, val bool) {
	mutex.Lock()
	pending[id] = val
	mutex.Unlock()
}

func testLiveness(id string) {
	stop := false
	defer func() {
		if stop {
			stopContainer(id)
		}
	}()

	// Inspect container to make sure it's running
	info, err := docker.InspectContainer(id)
	if err != nil {
		// It's kill, remove it from our state
		log.Println("WARN: Removing dead container", id[:16])
		stop = true
		return
	}
	if !info.State.Running {
		// Don't want to keep dead containers around
		log.Println("WARN: Removing non-running container", id[:16])
		stop = true
		return
	}

	// Allow the service to boot up - don't run liveness check for a minute
	if time.Since(info.State.StartedAt) < time.Minute {
		log.Println("INFO: Liveness: Skipping container", id[:16], "because it was launched recently.")
		return
	}

	ip := info.NetworkSettings.IPAddress
	url := "http://" + ip + ":" + servicePort + "/control/ping"
	client := http.Client{
		Timeout: timeout,
	}

	if pending[id] {
		log.Println("INFO: Liveness: check for", id[:16], "pending. Skipping.")
		return
	}

	setPending(id, true)
	res, err := client.Get(url)
	setPending(id, false)
	if res != nil {
		if res.StatusCode != 200 {
			log.Println("WARN: Liveness: check for container", id[:16], "returned code:", res.StatusCode, "(err", err, ")")

			stop = true
			return
		}
	}
	if err != nil {
		log.Println("WARN: Liveness: check for", id[:16], "failed with:", err)
		stop = true
		return
	}

	log.Println("INFO: Liveness: check for container", id[:16], " OK")
}

func monitor() {
	pending = make(map[string]bool)

	// Every `interval`, send a request to all the containers we're monitoring.
	for range time.Tick(interval) {
		containers, err := State.MonitoredContainers()
		if err != nil {
			continue
		}

		for _, container := range containers {
			go testLiveness(container)
		}
	}
}
