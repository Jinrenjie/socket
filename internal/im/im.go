package im

import (
	"fmt"

	"github.com/Jinrenjie/socket/database"
	"github.com/Jinrenjie/socket/internal/logs"
	"github.com/garyburd/redigo/redis"
)

type Status struct {
	Address  string `redis:"ip"`
	Platform string `redis:"platform"`
	Version  string `redis:"version"`
}

const (
	_prefix = "user"
)

func keyUser(id string) string {
	return _prefix + id
}

// Bind the user ID when the user goes online
func Online(id string, fd string, addr string, platform string, version string) {
	conn := database.Pool.Get()
	defer func() {
		if err := conn.Close(); err != nil {
			logs.OutPut("ERROR", "close-redis-connection", err.Error())
		}
	}()

	//HMSET website google www.google.com
	// key := fmt.Sprintf("users:%v", id)

	key := keyUser(id)
	value := fmt.Sprintf("%v-%v-%v", addr, platform, version)
	if _, err := conn.Do("HMSET", key, fd, value); err != nil {
		logs.OutPut("ERROR", "online", err.Error())
	}
	fmt.Println(key, fd, value)
}

// Unbind the user ID when the user goes offline
func Offline(id, fd string) {
	connection := database.Pool.Get()
	defer func() {
		if err := connection.Close(); err != nil {
			logs.OutPut("ERROR", "close-redis-connection", err.Error())
		}
	}()
	if _, err := connection.Do("HDEL", keyUser(id), fd); err != nil {
		logs.OutPut("ERROR", "offline", err.Error())
	}
}

// Check user status by id
func CheckById(id string) bool {
	connection := database.Pool.Get()
	defer func() {
		if err := connection.Close(); err != nil {
			logs.OutPut("ERROR", "close-redis-connection", err.Error())
		}
	}()
	r, err := redis.Bool(connection.Do("EXISTS", keyUser(id)))
	if err != nil {
		logs.OutPut("ERROR", "check-online", err.Error())
	}
	return r
}

func GetClients(id string) []string {
	connection := database.Pool.Get()
	defer func() {
		if err := connection.Close(); err != nil {
			logs.OutPut("ERROR", "close-redis-connection", err.Error())
		}
	}()
	clients, err := redis.Strings(connection.Do("HKEYS", keyUser(id)))
	if err != nil {
		logs.OutPut("ERROR", "get-clients", err.Error())
	}

	return clients
}
