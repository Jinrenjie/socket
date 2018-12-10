package database

import (
	"github.com/garyburd/redigo/redis"
	"log"
)

var (
	Pool *redis.Pool
)

func CreateRedisPool(addr, password string, db int) {
	Pool = &redis.Pool{
		MaxIdle:     30,
		MaxActive:   50,
		IdleTimeout: 5e9,
		Wait: true,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", addr,
				redis.DialDatabase(db),
				redis.DialPassword(password),
				redis.DialReadTimeout(10e9),
				redis.DialWriteTimeout(10e9),
				redis.DialConnectTimeout(10e9))
			if err != nil {
				log.Println(err)
				return nil, err
			}
			return c, err
		},
	}
}
