package main

import (
	"errors"
	"gopkg.in/redis.v3"
	"log"
)

const id = 4242424242

type state struct {
	Redis *redis.Client
	ID    int
}

var State state

func (s *state) Init() {
	// Connect to Redis
	s.Redis = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	pong, err := s.Redis.Ping().Result()
	if err != nil {
		log.Fatal("ERR: Redis says: ", err)
	}
	log.Println("Redis ping reply:", pong)

	// Set this instance's ID (TODO: Get from config/env)
	s.ID = id
}

func (s *state) Destroy() {
	s.Redis.Close()
	s.Redis = nil
}

func (s *state) AddContainer(id string) error {
	res := s.Redis.SAdd(key("monitored"), id)
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

func (s *state) RemoveContainer(id string) error {
	err := s.Redis.SRem(key("monitored"), id).Err()
	if err != nil {
		log.Println("removecontainer err:", err)
		State.Destroy()
		State.Init()

		return err
	}
	return err
}

func (s *state) MonitoredContainers() ([]string, error) {
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
