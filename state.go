package main

import (
	"log"
	dockerc "github.com/samalba/dockerclient"
	"gopkg.in/redis.v3"
	"errors"
)

const id = 4242424242

type state struct {
	Docker *dockerc.DockerClient
	Redis *redis.Client
	ID int
}

var State state

func(s *state) Init() {
	var err error

	// Connect to the Docker daemon
	s.Docker, err = dockerc.NewDockerClient("unix:////var/run/docker.sock", nil)
	if err != nil {
		log.Fatal("ERR: Docker says: ", err)
	}

	// Ensure we're actually connected
	info, err := s.Docker.Info()
	if err != nil {
		log.Fatal("ERR: Docker says: ", err)
	}
	log.Println("Running at", info.Name, "on", info.OperatingSystem)

	// Connect to Redis
	s.Redis = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		Password: "",
		DB: 0,
	})

	pong, err := s.Redis.Ping().Result()
	if err != nil {
		log.Fatal("ERR: Redis says: ", err)
	}
	log.Println("Redis ping reply:", pong)

	// Set this instance's ID (TODO: Get from config/env)
	s.ID = id
}

func(s *state) Destroy() {
	s.Redis.Close()
	s.Redis = nil
	s.Docker = nil
}

// TODO: Refactor below into redis.go
func(s *state) AddContainer(id string, ip string) error {
	res := s.Redis.SAdd(key("monitored"), id + ":" + ip)
	if res == nil {
		State.Destroy()
		State.Init()
		return errors.New("No reply from Redis")
	}

	err := res.Err()
	if err != nil {
		log.Println("addcontainer err:", err)
		State.Destroy()
		State.Init()

		return err
	}
	return err
}

func(s *state) RemoveContainer(id string, ip string) error {
	err := s.Redis.SRem(key("monitored"), id + ":" + ip).Err()
	if err != nil {
		log.Println("removecontainer err:", err)
		State.Destroy()
		State.Init()

		return err
	}
	return err
}

func(s *state) MonitoredContainers() ([]string, error) {
	res := s.Redis.SMembers(key("monitored"))
	if res == nil {
		State.Destroy()
		State.Init()
		return nil, errors.New("No reply from Redis")
	}

	ret, err := res.Result()
	if err != nil {
		log.Println("error from redis:", err)
		State.Destroy()
		State.Init()
		return nil, err
	}
	return ret, err
}
