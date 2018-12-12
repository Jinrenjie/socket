package im

import (
	"fmt"
	"log"
	"time"

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
			logs.Save(&logs.Payload{
				Uid:        id,
				Fd:         fd,
				Type:       "close-redis-connection",
				Body:       err.Error(),
				CreateTime: time.Now().Unix(),
				CreateDate: time.Now().Format("2006-01-02"),
			})
		}
	}()

	//HMSET website google www.google.com
	// key := fmt.Sprintf("users:%v", id)

	key := keyUser(id)
	value := fmt.Sprintf("%v-%v-%v", addr, platform, version)
	if _, err := conn.Do("HMSET", key, fd, value); err != nil {
		log.Fatalf("connection.Do HMSET error %v", err)
		logs.Save(&logs.Payload{
			Uid:        id,
			Fd:         fd,
			Type:       "online",
			Body:       err.Error(),
			CreateTime: time.Now().Unix(),
			CreateDate: time.Now().Format("2006-01-02"),
		})
	}
	fmt.Println(key, fd, value)
}

// Unbind the user ID when the user goes offline
func Offline(id, fd string) {
	connection := database.Pool.Get()
	defer func() {
		if err := connection.Close(); err != nil {
			logs.Save(&logs.Payload{
				Uid:        id,
				Fd:         fd,
				Type:       "close-redis-connection",
				Body:       err.Error(),
				CreateTime: time.Now().Unix(),
				CreateDate: time.Now().Format("2006-01-02"),
			})
		}
	}()
	if _, err := connection.Do("HDEL", keyUser(id), fd); err != nil {
		logs.Save(&logs.Payload{
			Uid:        id,
			Fd:         fd,
			Type:       "offline",
			Body:       err.Error(),
			CreateTime: time.Now().Unix(),
			CreateDate: time.Now().Format("2006-01-02"),
		})
	}
}

// Check user status by id
func CheckById(id string) bool {
	connection := database.Pool.Get()
	defer func() {
		if err := connection.Close(); err != nil {
			logs.Save(&logs.Payload{
				Uid:        id,
				Fd:         "",
				Type:       "close-redis-connection",
				Body:       err.Error(),
				CreateTime: time.Now().Unix(),
				CreateDate: time.Now().Format("2006-01-02"),
			})
		}
	}()
	r, err := redis.Bool(connection.Do("EXISTS", keyUser(id)))
	if err != nil {
		logs.Save(&logs.Payload{
			Uid:        id,
			Fd:         "",
			Type:       "check-online",
			Body:       err.Error(),
			CreateTime: time.Now().Unix(),
			CreateDate: time.Now().Format("2006-01-02"),
		})
	}
	return r
}

func GetClients(id string) []string {
	connection := database.Pool.Get()
	defer func() {
		if err := connection.Close(); err != nil {
			logs.Save(&logs.Payload{
				Uid:        id,
				Fd:         "",
				Type:       "close-redis-connection",
				Body:       err.Error(),
				CreateTime: time.Now().Unix(),
				CreateDate: time.Now().Format("2006-01-02"),
			})
		}
	}()
	clients, err := redis.Strings(connection.Do("HKEYS", keyUser(id)))
	if err != nil {
		logs.Save(&logs.Payload{
			Uid:        id,
			Fd:         "",
			Type:       "get-clients",
			Body:       err.Error(),
			CreateTime: time.Now().Unix(),
			CreateDate: time.Now().Format("2006-01-02"),
		})
	}

	return clients
}
