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

	res, err := http.Get(url)
	if err != nil || res.StatusCode != 200 {
		log.Println("WARN: Liveness check for container", id[:16], "returned code:", res.StatusCode, "(err", err, ")")

		// Stop the container
		stopContainer(id)
		State.RemoveContainer(id, ip)
		return
	}

	log.Println("INFO: Liveness check for container", id[:16], "is OK.")
}

func monitor() {
	// Every `interval`, send a request to all the containers we're monitoring.
	for range time.Tick(interval) {
		containers, err := State.MonitoredContainers()
		if err != nil {
			log.Fatal("Error getting monitored containers: ", err)
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
