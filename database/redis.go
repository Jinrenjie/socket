package database

import (
	"github.com/garyburd/redigo/redis"
	"log"
)

var (
	Pool *redis.Pool
)

func CreateRedisPool(addr, password string) {
	Pool = &redis.Pool{
		MaxIdle:     100,
		MaxActive:   500,
		IdleTimeout: 5e9,
		Wait: true,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", addr,
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
