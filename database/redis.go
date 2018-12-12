package database

import (
	"log"
	"time"

	"github.com/garyburd/redigo/redis"
)

var (
	Pool *redis.Pool
)

func CreateRedisPool(addr, pwd string, db int) {
	Pool = &redis.Pool{
		MaxIdle:     5,               //定义连接池中最大连接数（超过这个数会关闭老的链接，总会保持这个数）
		MaxActive:   20,              //最大的激活连接数，表示同时最多有N个连接
		IdleTimeout: 5 * time.Second, //定义链接的超时时间，每次p.Get()的时候会检测这个连接是否超时（超时会关闭，并释放可用连接数）
		Wait:        true,            // 当可用连接数为0是，那么当wait=true,那么当调用p.Get()时，会阻塞等待，否则，返回nil.
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", addr,
				redis.DialDatabase(db),
				redis.DialPassword(pwd),
				redis.DialReadTimeout(10*time.Second),
				redis.DialWriteTimeout(10*time.Second),
				redis.DialConnectTimeout(10*time.Second))
			if err != nil {
				c.Close()
				log.Fatalf("redis.Dial error(%v)\n", err)
				return nil, err
			}
			// if _, err := c.Do("AUTH", pwd); err != nil {
			// 	c.Close()
			// 	log.Fatalf("c.Do auth error(%v)\n", err)
			// 	return nil, err
			// }
			c.Do("SELECT", db) //选择db
			return c, err
		},
		// 如果设置了给func,那么每次p.Get()的时候都会调用改方法来验证连接的可用性
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}
}
