package im

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"socket/database"
	"socket/internal/logs"
	"time"
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
	key := fmt.Sprintf("users:%v", id)
	value := fmt.Sprintf("%v-%v-%v", addr, platform, version)
	if _, err := connection.Do("HMSET", key, fd, value); err != nil {
		logs.Save(&logs.Payload{
			Uid:        id,
			Fd:         fd,
			Type:       "online",
			Body:       err.Error(),
			CreateTime: time.Now().Unix(),
			CreateDate: time.Now().Format("2006-01-02"),
		})
	}
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
	key := fmt.Sprintf("users:%v", id)
	if _, err := connection.Do("HDEL", key, fd); err != nil {
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
	key := fmt.Sprintf("users:%v", id)
	r, err := redis.Bool(connection.Do("EXISTS", key))
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
	key := fmt.Sprintf("users:%v", id)
	clients, err := redis.Strings(connection.Do("HKEYS", key))
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