package main

import (
	"log"
	"net/http"
	"sync"
	"time"
	"io/ioutil"
)

// Prime numbers so they don't coincide
const interval = 7 * time.Second
const timeout = 13 * time.Second
const gracePeriod = 20 * time.Second

var pending map[string]bool
var mutex sync.Mutex

func setPending(id string, val bool) {
	mutex.Lock()
	pending[id] = val
	mutex.Unlock()
}

func testLiveness(id string, destructive bool) ContainerStatus {
	stop := false
	defer func() {
		if stop && destructive {
			stopContainer(id)
		}
	}()

	// Inspect container to make sure it's running
	info, err := docker.InspectContainer(id)
	if err != nil {
		// It's kill, remove it from our state
		log.Println("WARN: container", id[:16], "is dead.")
		stop = true
		return DEAD
	}
	// Allow the service to boot up - don't run liveness check for a minute
	if time.Since(info.State.StartedAt) < gracePeriod && destructive {
		log.Println("INFO: Liveness: Skipping container", id[:16], "because it was launched recently.")
		return STARTING
	}
	if !info.State.Running {
		// Don't want to keep dead containers around
		log.Println("WARN: container", id[:16], "is not running.")
		stop = true
		return NOT_RUNNING
	}

	ip := info.NetworkSettings.IPAddress
	url := "http://" + ip + ":" + servicePort + "/control/state/" + id
	client := http.Client{
		Timeout: timeout,
	}

	if pending[id] {
		log.Println("INFO: Liveness: check for", id[:16], "pending. Skipping.")
		return RUNNING
	}

	setPending(id, true)
	res, err := client.Get(url)
	setPending(id, false)
	if res != nil {
		if res.StatusCode != 200 {
			log.Println("WARN: Liveness: check for container", id[:16], "returned code:", res.StatusCode, "(err", err, ")")

			stop = true
			return DEAD
		}
	}
	if err != nil {
		log.Println("WARN: Liveness: check for", id[:16], "failed with:", err)
		stop = true
		return DEAD
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println("WARN: Liveness: check for", id[:16], "failed with:", err)
		stop = true
		return DEAD
	}

	if string(body) == "disconnected" {
		stop = true
		log.Println("INFO: Liveness: client d/c, killing", id[:16])
		return NOT_RUNNING
	}

	log.Println("INFO: Liveness: check for container", id[:16], " OK")
	return READY
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
			go testLiveness(container, true)
		}
	}
}
