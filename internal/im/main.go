package im

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"github.com/Shopify/sarama"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"socket/internal/logs"
	"strings"
	"sync"
	"time"
)

type DeliverResult struct {
	Fd string `json:"fd"`
	Status string `json:"status"`
}

var (
	// Define a broker and topic for kafka
	broker, topic string
	producer sarama.AsyncProducer
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
		return "", "", nil, err
	}

	// Upgrade initial GET request to a websocket
	connection, err := upgrader.Upgrade(response, request, nil)
	if err != nil {
		log.Fatal(err, 2)
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
		tokenStr := request.URL.String()[index + 6:]
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
		platform = "iOS"
	}
	if id == "" || version == "" || platform == "" {
		return id, version, platform, errors.New("authentication failure")
	}
	return id, version, platform, nil
}

// Connections handler logic
func Handle(response http.ResponseWriter, request *http.Request) {
	id, fd, connection, err := upgrade(response, request)
	if err != nil {
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

	//logGenerate(&logs.Payload{
	//	Uid:        id,
	//	Fd:         fd,
	//	Type:       "connection",
	//	Body:       "Connected",
	//	CreateTime: time.Now().String(),
	//	CreateDate: time.Now().Format("2006-01-02"),
	//	Microtime:  time.Now().UnixNano() / 1000,
	//})

	// Register our new client
	connections.Store(fd, connection)

	// Set read dead line
	if err := connection.SetReadDeadline(time.Now().Add(120e9)); err != nil {
		// TODO handle err
	}
	connection.SetPingHandler(func(appData string) error {
		if err := connection.SetReadDeadline(time.Now().Add(120e9)); err != nil {
			// TODO handle err
		}
		return nil
	})

	for {
		var msg Payload
		// Read in a new message as JSON and map it to a Message object
		if err := connection.ReadJSON(&msg); err != nil {
			//logGenerate(&logs.Payload{
			//	Uid:        id,
			//	Fd:         fd,
			//	Type:       "connection",
			//	Body:       "Disconnected",
			//	CreateTime: time.Now().String(),
			//	CreateDate: time.Now().Format("2006-01-02"),
			//	Microtime:  time.Now().UnixNano() / 1000,
			//})
			connections.Delete(fd)
			Offline(id, fd)
			break
		}

		if err := connection.SetReadDeadline(time.Now().Add(120e9)); err != nil {
			log.Printf("set read dead line error %v", err)
		}

		DeliverMessage(msg.Body.To, msg)
	}
}

// Log push to kafka producer
func logGenerate(content *logs.Payload) {
	if broker == "" && topic == "" {
		kafka := viper.GetStringMapString("kafka")
		broker = kafka["broker"]
		topic = kafka["topic"]
	}
	if producer == nil {
		producer = logs.CreateAsyncProducer(broker)
	}
	producer.Input() <- &sarama.ProducerMessage{
		Topic: topic,
		Key: sarama.StringEncoder(content.Uid),
		Value: content,
	}
}

// Deliver message to client
func DeliverMessage(id string, message Payload) []DeliverResult {
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
			if err := connection.WriteJSON(message); err != nil {
				client.Status = "failure"
				log.Printf("send message error %v", err)
			} else {
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
