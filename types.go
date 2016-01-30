package main

type StartResponse struct {
	Endpoint    string
	ContainerID string
}

type GenericResponse struct {
	Success bool
}

type StatusMessageResponse struct {
	Success bool
	Message string
}

type ContainerStatus int
const (
	DEAD ContainerStatus = iota
	NOT_RUNNING
	RUNNING
	STARTING
	READY
)
//go:generate stringer -type=ContainerStatus

type StatusResponse struct {
	ContainerID string
	Status string
}
