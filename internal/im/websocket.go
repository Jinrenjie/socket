package im

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/Jinrenjie/socket/internal/logs"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/naoina/denco"
)

type DeliverResult struct {
	Fd     string `json:"fd"`
	Status string `json:"status"`
}

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  4096,
		WriteBufferSize: 4096,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	// Define a map to hold the WebSocket connection
	connections sync.Map
)

// Web Socket upgrade
func upgrade(response http.ResponseWriter, request *http.Request) (string, string, *websocket.Conn, error) {
	id, version, platform, err := bind(request)
	if err != nil {
		log.Printf("bind error %v", err)
		return "", "", nil, err
	}

	// Upgrade initial GET request to a websocket
	connection, err := upgrader.Upgrade(response, request, nil)
	if err != nil {
		log.Fatalf("upgrader.Upgrade %v", err)
		return "", "", nil, err
	}

	fd := uuid.New().String()
	Online(id, fd, connection.RemoteAddr().String(), platform, version)
	return id, fd, connection, nil
}

// Bind user id to connection
func bind(request *http.Request) (id, version, platform string, err error) {
	params := request.URL.Query()
	if request.URL.String() == "" {
		return "", "", "", errors.New("query string parse fails")
	}
	index := strings.LastIndex(request.URL.String(), "token=")
	if index > 0 {
		tokenStr := request.URL.String()[index+6:]
		if tokenStr != "" {
			tokenValue, err := url.ParseQuery(tokenStr)
			if err != nil {
				return "", "", "", errors.New("query string parse fails")
			}
			hash := md5.New()
			hash.Write([]byte(tokenValue.Get("oauth2") + tokenValue.Get("timestamp")))
			cipherHexStr := hash.Sum(nil)
			if tokenValue.Get("signature") != hex.EncodeToString(cipherHexStr) {
				return "", "", "", errors.New("token validation fails")
			}
		}
	}
	id = params.Get("id")
	version = params.Get("version")
	if version == "" {
		version = "1.0.1"
	}
	platform = params.Get("platform")
	if platform == "" {
		platform = "web"
	}
	if id == "" || version == "" || platform == "" {
		return id, version, platform, errors.New("authentication failure")
	}
	return id, version, platform, nil
}

// Connections handler logic
func Handle(response http.ResponseWriter, request *http.Request, params denco.Params) {
	id, fd, connection, err := upgrade(response, request)
	if err != nil {
		log.Printf("upgrade error %v", err)
		response.WriteHeader(422)
		return
	}
	// Make sure we close the connection when the function returns
	defer func() {
		if connection != nil {
			if err := connection.Close(); err != nil {
				log.Printf("close connection error %v", err)
			}
		}
	}()

	logs.Save(&logs.Payload{
		Uid:        id,
		Fd:         fd,
		Type:       "connection",
		Body:       "Connected",
		CreateTime: time.Now().Unix(),
		CreateDate: time.Now().Format("2006-01-02"),
	})

	// Register our new client
	connections.Store(fd, connection)

	// Set read dead line
	//if err := connection.SetReadDeadline(time.Now().Add(120e9)); err != nil {
	//	// TODO handle err
	//}
	//connection.SetPingHandler(func(appData string) error {
	//	if err := connection.SetReadDeadline(time.Now().Add(120e9)); err != nil {
	//		// TODO handle err
	//	}
	//	return nil
	//})

	for {
		// Read in a new message as JSON and map it to a Message object
		_, _, err := connection.ReadMessage()
		if err != nil {
			logs.OutPut(id, fd, "connection", "Disconnected")
			connections.Delete(fd)
			// Offline(id, fd)
			break
		}

		//if err := connection.SetReadDeadline(time.Now().Add(120e9)); err != nil {
		//	log.Printf("set read dead line error %v", err)
		//}
	}
}

// Deliver message to client
func DeliverMessage(id string, message []byte) []DeliverResult {
	clients := GetClients(id)

	result := make([]DeliverResult, 0)

	for _, fd := range clients {
		oringin, ok := connections.Load(fd)
		var client = DeliverResult{
			Fd: fd,
		}
		if ok {
			contype := reflect.ValueOf(oringin)
			connection := contype.Interface().(*websocket.Conn)
			if err := connection.WriteMessage(websocket.TextMessage, message); err != nil {
				client.Status = "failure"
				log.Printf("send message error %v", err)
				logs.OutPut(id, fd, "deliver", err.Error())
			} else {
				logs.OutPut(id, fd, "deliver", "success")
				client.Status = "success"
			}
			result = append(result, client)
		} else {
			client.Status = "failure"
			result = append(result, client)
		}
	}
	return result
}
