package api

import (
	"encoding/json"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/naoina/denco"
	"io/ioutil"
	"log"
	"net/http"
	"socket/database"
	"socket/internal/im"
	"strings"
)

type Response struct {
	Status int `json:"status"`
	Message string `json:"message"`
	Data interface{} `json:"data"`
}

// Deliver message to user
func Deliver(writer http.ResponseWriter, request *http.Request, params denco.Params)  {
	res := Response{
		Status: 200,
		Message: "success",
	}

	bytes, err := ioutil.ReadAll(request.Body)
	if err != nil {
		fmt.Println("ERR: read request ", err)
	}

	if len(bytes) == 0 {
		res.Status = 422
		res.Message = "request body is null"
		res.Data = make([]int, 0)
		writer.WriteHeader(422)
	}

	id := params.Get("id")
	if res.Status == 200 {
		online := im.CheckById(id)

		if online {
			res.Data = im.DeliverMessage(id, bytes)
		} else {
			res.Status = 200
			res.Message = "this user is offline"
			res.Data = make([]int, 0)
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
func CheckOnline(writer http.ResponseWriter, request *http.Request, params denco.Params)  {
	res := Response{
		Status: 200,
		Message: "success",
	}

	id := params.Get("id")
	res.Data = im.CheckById(id)

	result, err := json.Marshal(res)
	if err != nil {
		fmt.Println(err)
	}
	if _, err := writer.Write(result); err != nil {
		log.Fatal("ERR: marshal request", err)
	}
}

// Get all connections
func Connections(writer http.ResponseWriter, request *http.Request, params denco.Params)  {
	type User struct {
		Id string `json:"id"`
		Clients interface{} `json:"clients"`
	}

	type Client struct {
		Fd string `json:"fd"`
		Address string `json:"address"`
		Platform string `json:"platform"`
		Version string `json:"version"`
	}
	var (
		userskey []string
		err error
	)
	connection := database.Pool.Get()
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
			Id: temp[1],
		}

		clients, err := redis.StringMap(connection.Do("HGETALL", key))
		if err != nil {
		}
		uclients := make([]Client, len(clients))
		j := 0
		for fd, infostring := range clients{
			info := strings.Split(infostring, "-")
			client := Client{
				Fd: fd,
				Address: info[0],
				Platform: info[1],
				Version: info[2],
			}
			uclients[j] = client
			j++

		}
		user.Clients = uclients
		users[i] = user
		i++
	}

	if result, err := json.Marshal(Response{
		Status: 200,
		Message: "success",
		Data: users,
	}); err != nil {

	} else {
		if _, err := writer.Write(result); err != nil {
			log.Fatal("ERR: marshal request", err)
		}
	}
}
