package api

import (
	"encoding/json"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/spf13/viper"
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

func Register()  {
	apiConf := viper.GetStringMapString("api")
	prefix := apiConf["prefix"]

	if prefix == "" {
		prefix = "/api"
	}

	// API
	http.HandleFunc(prefix, func(writer http.ResponseWriter, request *http.Request) {
		data := make([]string, 0)
		response, err := json.Marshal(&Response{
			Status: 200,
			Message: "Success",
			Data: data,
		})
		if err != nil {
			log.Fatal(err)
		}
		if _, err := writer.Write(response); err != nil {
			log.Fatal(err)
		}
	})
	// Deliver message
	http.HandleFunc(prefix + "/deliver", func(writer http.ResponseWriter, request *http.Request) {
		res := Response{
			Status: 200,
			Message: "success",
		}

		if request.Method != "POST" {
			res.Status = 405
			res.Message = "Method Not Allowed"
			res.Data = false
			writer.WriteHeader(405)
		}

		if request.Header.Get("Content-Type") != "application/json" {
			res.Status = 400
			res.Message = "Content-type must be application/json"
			res.Data = false
			writer.WriteHeader(400)
		}

		bytes, err := ioutil.ReadAll(request.Body)
		if err != nil {
			fmt.Println("ERR: read request ", err)
		}

		if len(bytes) == 0 {
			res.Status = 422
			res.Message = "request body is null"
			res.Data = false
			writer.WriteHeader(422)
		}

		if res.Status == 200 {
			msg := im.Payload{}

			if err := json.Unmarshal(bytes, &msg); err != nil {
				log.Fatal("ERR: unmarshal request", err)
			}
			online := im.CheckById(msg.Body.To)

			if online {
				res.Data = im.DeliverMessage(msg.Body.To, msg)
			} else {
				res.Status = 404
				res.Message = "The user is not online"
			}
		}

		result, err := json.Marshal(&res)
		if err != nil {
			log.Fatal("ERR: marshal request", err)
		}

		if _, err := writer.Write(result); err != nil {
			log.Fatal("ERR: marshal request", err)
		}
	})
	// Check online
	http.HandleFunc(prefix + "/check", func(writer http.ResponseWriter, request *http.Request) {
		res := Response{
			Status: 200,
			Message: "success",
		}

		if request.Method != "POST" {
			res.Status = 405
			res.Message = "Method Not Allowed"
			res.Data = false
			writer.WriteHeader(405)
		}

		if request.Header.Get("Content-Type") != "application/json" {
			res.Status = 400
			res.Message = "Content-type must be application/json"
			res.Data = false
			writer.WriteHeader(400)
		}

		bytes, err := ioutil.ReadAll(request.Body)
		if err != nil {
			fmt.Println("ERR: read request ", err)
		}

		if len(bytes) == 0 {
			res.Status = 422
			res.Message = "request body is null"
			res.Data = false
			writer.WriteHeader(422)
		}

		if res.Status == 200 {
			var form = struct {
				Id string
			}{}

			err = json.Unmarshal(bytes, &form)
			if err != nil {
				fmt.Println("ERR: unmarshal request ", err)
			}

			res.Data = im.CheckById(form.Id)
		}

		result, err := json.Marshal(res)
		if err != nil {
			fmt.Println(err)
		}
		if _, err := writer.Write(result); err != nil {
			log.Fatal("ERR: marshal request", err)
		}
	})

	// Get online connections
	http.HandleFunc(prefix + "/connections", func(writer http.ResponseWriter, request *http.Request) {
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

		connection := database.Pool.Get()
		userskey, err := redis.Strings(connection.Do("KEYS", "users:*"))
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
	})
}
