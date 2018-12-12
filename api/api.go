package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/Jinrenjie/socket/database"
	"github.com/Jinrenjie/socket/internal/im"
	"github.com/Jinrenjie/socket/internal/logs"
	"github.com/garyburd/redigo/redis"
	"github.com/naoina/denco"
)

type Response struct {
	Code   int         `json:"code"`
	Msg    string      `json:"msg"`
	Result interface{} `json:"result"`
}

type User struct {
	Uid     string      `json:"uid"`
	Clients interface{} `json:"clients"`
}

type Client struct {
	Fd       string `json:"fd"`
	Address  string `json:"address"`
	Platform string `json:"platform"`
	Version  string `json:"version"`
}

// Deliver message to user
func Deliver(writer http.ResponseWriter, request *http.Request, params denco.Params) {
	res := Response{
		Code: 0,
		Msg:  "success",
	}

	bytes, err := ioutil.ReadAll(request.Body)
	if err != nil {
		fmt.Println("ERR: read request ", err)
	}

	if len(bytes) == 0 {
		res.Code = 422
		res.Msg = "request body is null"
		res.Result = make([]int, 0)
		writer.WriteHeader(422)
	}

	id := params.Get("id")
	if res.Code == 0 {
		online := im.CheckById(id)

		if online {
			res.Result = im.DeliverMessage(id, bytes)
		} else {
			res.Code = 0
			res.Msg = "this user is offline"
			res.Result = make([]int, 0)
		}
	}

	result, err := json.Marshal(&res)
	if err != nil {
		log.Fatal("ERR: marshal request", err)
	}

	if _, err := writer.Write(result); err != nil {
		log.Fatal("ERR: marshal request", err)
	}
}

// Check user online status
func CheckOnline(writer http.ResponseWriter, request *http.Request, params denco.Params) {
	res := Response{
		Code: 0,
		Msg:  "success",
	}

	id := params.Get("id")
	res.Result = im.CheckById(id)

	result, err := json.Marshal(res)
	if err != nil {
		fmt.Println(err)
	}
	if _, err := writer.Write(result); err != nil {
		log.Fatal("ERR: marshal request", err)
	}
}

// Get all connections
func Connections(writer http.ResponseWriter, request *http.Request, params denco.Params) {
	var (
		userskey []string
		err      error
	)
	connection := database.Pool.Get()
	defer func() {
		if err := connection.Close(); err != nil {
			logs.OutPut("api", "api", "close-redis-connection", err.Error())
		}
	}()
	origin := request.URL.Query()
	id := origin.Get("id")
	if id != "" {
		userskey, err = redis.Strings(connection.Do("KEYS", fmt.Sprintf("users:%v", id)))
	} else {
		userskey, err = redis.Strings(connection.Do("KEYS", "users:*"))
	}

	if err != nil {
		fmt.Println(err)
	}

	users := make([]User, len(userskey))

	i := 0
	for _, key := range userskey {
		temp := strings.Split(key, ":")
		user := User{
			Uid: temp[1],
		}

		clients, err := redis.StringMap(connection.Do("HGETALL", key))
		if err != nil {
		}
		uclients := make([]Client, len(clients))
		j := 0
		for fd, infostring := range clients {
			info := strings.Split(infostring, "-")
			client := Client{
				Fd:       fd,
				Address:  info[0],
				Platform: info[1],
				Version:  info[2],
			}
			uclients[j] = client
			j++

		}
		user.Clients = uclients
		users[i] = user
		i++
	}

	if result, err := json.Marshal(Response{
		Code:   0,
		Msg:    "success",
		Result: users,
	}); err != nil {

	} else {
		if _, err := writer.Write(result); err != nil {
			log.Fatal("ERR: marshal request", err)
		}
	}
}

func Health(writer http.ResponseWriter, request *http.Request, params denco.Params) {
	writer.WriteHeader(200)
	if _, err := writer.Write([]byte("Socket Service is OK!")); err != nil {
		log.Printf("%v", err)
	}
}
