package database

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/spf13/viper"
	mgo "gopkg.in/mgo.v2"
)

var (
	dialInfo *mgo.DialInfo
	err      error
	Session  *mgo.Session
	once     sync.Once
)

func Connection() *mgo.Session {
	logger := viper.GetStringMapString("log")
	addr := fmt.Sprintf("%v:%v", logger["host"], logger["port"])

	dialInfo = &mgo.DialInfo{
		Addrs:     []string{addr},
		Direct:    false,
		Timeout:   time.Second * 80,
		Database:  logger["database"],
		Source:    "admin",
		Username:  logger["username"],
		Password:  logger["password"],
		PoolLimit: 4096,
	}

	once.Do(func() {
		Session, err = mgo.DialWithInfo(dialInfo)
		if err != nil {
			Session.Close()
			log.Printf("%v", err)
		}

		Session.SetMode(mgo.Monotonic, true)
	})

	return Session.Clone()
}
