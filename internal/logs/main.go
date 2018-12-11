package logs

import (
	"encoding/json"
	"log"

	"github.com/Shopify/sarama"
	"github.com/fengpf/socket/database"
	"github.com/spf13/viper"
)

type Payload struct {
	Uid        string      `json:"uid"`
	Fd         string      `json:"fd"`
	Type       string      `json:"type"`
	Body       interface{} `json:"body"`
	CreateTime int64       `json:"create_time"`
	CreateDate string      `json:"create_date"`
	encoded    []byte
	err        error
}

var (
	// Define a broker and topic for kafka
	broker, topic string
	producer      sarama.AsyncProducer
)

func (ale *Payload) ensureEncoded() {
	if ale.encoded == nil && ale.err == nil {
		ale.encoded, ale.err = json.Marshal(ale)
	}
}

func (ale *Payload) Length() int {
	ale.ensureEncoded()
	return len(ale.encoded)
}

func (ale *Payload) Encode() ([]byte, error) {
	ale.ensureEncoded()
	return ale.encoded, ale.err
}

func createAsyncProducer(broker string) sarama.AsyncProducer {
	config := sarama.NewConfig()

	config.Producer.RequiredAcks = sarama.WaitForLocal
	config.Producer.Compression = sarama.CompressionSnappy
	config.Producer.Flush.Frequency = 3e9
	config.Producer.Timeout = 3e9

	producer, err := sarama.NewAsyncProducer([]string{broker}, config)

	if err != nil {
		log.Fatalln("Failed to start Sarama producer:", err)
	}

	// We will just log to STDOUT if we're not able to produce messages.
	// Note: messages will only be returned here after all retry attempts are exhausted.
	go func() {
		for err := range producer.Errors() {
			log.Println("Failed to write access log entry:", err)
		}
	}()
	return producer
}

func Save(payload *Payload) {
	//logKafka(payload)
	logMongoDB(payload)
}

func logMongoDB(payload *Payload) {
	logger := viper.GetStringMapString("log")
	session := database.Connection()
	collection := session.DB(logger["database"]).C(logger["collection"])

	defer func() {
		session.Close()
	}()

	if err := collection.Insert(payload); err != nil {

	}
}

// Log push to kafka producer
func logKafka(content *Payload) {
	if broker == "" && topic == "" {
		kafka := viper.GetStringMapString("kafka")
		broker = kafka["broker"]
		topic = kafka["topic"]
	}
	if producer == nil {
		producer = createAsyncProducer(broker)
	}
	producer.Input() <- &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(content.Uid),
		Value: content,
	}
}
