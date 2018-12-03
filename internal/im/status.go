package im

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"log"
	"socket/database"
)

type Status struct {
	Address string `redis:"ip"`
	Platform    string `redis:"platform"`
	Version     string `redis:"version"`
}

// Bind the user ID when the user goes online
func Online(id string, fd string, addr string, platform string, version string) {
	connection := database.Pool.Get()
	defer func() {
		err := connection.Close()
		if err != nil {
			log.Printf("%v", err)
		}
	}()
	key := fmt.Sprintf("users:%v", id)
	value := fmt.Sprintf("%v-%v-%v", addr, platform, version)
	_, err := connection.Do("HMSET", key, fd, value)
	if err != nil {
		log.Printf("%v", err)
	}
	err = connection.Close()
	if err != nil {
		log.Printf("%v", err)
	}
}

// Unbind the user ID when the user goes offline
func Offline(id, fd string) {
	log.Printf("close: %v:%v", id, fd)
	connection := database.Pool.Get()
	defer func() {
		err := connection.Close()
		if err != nil {
			log.Printf("%v", err)
		}
	}()
	key := fmt.Sprintf("users:%v", id)
	_, err := connection.Do("HDEL", key, fd)
	if err != nil {
		log.Printf("User offline %v", err)
	}
	err = connection.Close()
	if err != nil {
		log.Printf("%v", err)
	}
}

// Check user status by id
func CheckById(id string) bool {
	connection := database.Pool.Get()
	key := fmt.Sprintf("users:%v", id)
	r, err := redis.Bool(connection.Do("EXISTS", key))
	if err != nil {
		log.Println(err)
	}
	return r
}

func GetClients(id string) []string {
	connection := database.Pool.Get()
	key := fmt.Sprintf("users:%v", id)
	clients, err := redis.Strings(connection.Do("HKEYS", key))
	if err != nil {
		log.Println(err)
	}

	return clients
}