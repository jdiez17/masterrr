package main

type StartResponse struct {
	Endpoint string
	ContainerID string
}

type GenericResponse struct {
	Success bool
}

type StatusMessageResponse struct {
	Success bool
	Message string
}
