package logs

import (
	"encoding/json"
	"github.com/Shopify/sarama"
	"log"
)

type Payload struct {
	Uid string `json:"uid"`
	Fd string `json:"fd"`
	Type string `json:"type"`
	Body string `json:"body"`
	CreateTime string `json:"create_time"`
	CreateDate string `json:"create_date"`
	Microtime int64 `json:"microtime"`
	encoded []byte
	err     error
}

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

func CreateAsyncProducer(broker string) sarama.AsyncProducer {
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