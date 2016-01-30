package main

import (
	"net/http"
	"log"
	"strings"
)

const servicePort = "8081"
const port = "8082"
const image = "jdiez/dockerrr"

// Rate limit by ip!
func startHandler(r *http.Request) (interface{}, HTTPError) {
	// Spin up a container
	id, err := startContainer(image)
	if err != nil {
		return nil, err
	}
	ip, err := getContainerIP(id)
	if err != nil {
		return nil, err
	}

	// Add to the list of containers managed by this process
	if err := State.AddContainer(id, ip); err != nil {
		stopContainer(id)
		return nil, NewHTTPError("Unable to add container to Redis.", 500)
	}

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
	ip, err := getContainerIP(id)
	if err != nil {
		return nil, err
	}
	if err := stopContainer(id); err != nil {
		return nil, err
	}

	// Remove the list of containers managed by this process
	if err := State.RemoveContainer(id, ip); err != nil {
		return nil, NewHTTPError("Unable to remove container from Redis.", 500)
	}

	return GenericResponse{Success: true}, nil
}

func main() {
	State.Init()
	go monitor()

	http.HandleFunc("/start/", jsonM(startHandler))
	http.HandleFunc("/stop/", jsonM(stopHandler))
	log.Println("masterrr: listening on", ":" + port)
	http.ListenAndServe(":" + port, nil)
}
