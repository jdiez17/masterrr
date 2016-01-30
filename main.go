package main

import (
	"log"
	"net/http"
	"strings"
)

const servicePort = "8081"
const port = "8082"
const image = "jdiez/dockerrr"

// Rate limit by ip!
func startHandler(r *http.Request) (interface{}, HTTPError) {
	// Spin up a container
	info, err := startContainer(image)
	if err != nil {
		return nil, err
	}

	return StartResponse{
		Endpoint:    info.IP + ":" + servicePort,
		ContainerID: info.ID,
	}, nil
}

func stopHandler(r *http.Request) (interface{}, HTTPError) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 3 {
		return nil, NewHTTPError("ID argument not given.", 400)
	}
	id := parts[2]
	if len(id) < 30 {
		return nil, NewHTTPError("Invalid ID", 400)
	}

	if err := stopContainer(id); err != nil {
		return nil, err
	}
	return GenericResponse{Success: true}, nil
}

func statusHandler(r *http.Request) (interface{}, HTTPError) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 3 {
		return nil, NewHTTPError("ID argument not given.", 400)
	}
	id := parts[2]
	if len(id) < 30 {
		return nil, NewHTTPError("Invalid ID", 400)
	}

	return StatusResponse{ContainerID: id, Status: testLiveness(id, false).String()}, nil
}

func main() {
	dockerConnect()
	State.Init()
	go monitor()

	http.HandleFunc("/start/", allowCrossOrigin(jsonM(startHandler)))
	http.HandleFunc("/stop/",  allowCrossOrigin(jsonM(stopHandler)))
	http.HandleFunc("/status/",allowCrossOrigin(jsonM(statusHandler)))
	log.Println("masterrr: listening on", ":"+port)
	http.ListenAndServe(":"+port, nil)
}
