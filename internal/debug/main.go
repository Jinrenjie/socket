package debug

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	_ "github.com/mkevac/debugcharts"
)

func StartDebug() {
	if err := http.ListenAndServe(":8088", nil); err != nil {
		log.Println(err)
	}
}

var (
	help bool
	n    int
	p    string
	t    int64
)

func init() {
	flag.BoolVar(&help, "help", false, "帮助")
	flag.IntVar(&n, "n", 1000, "设置客户端并发数量")
	flag.StringVar(&p, "p", "", "设置用于测试的消息内容")
	flag.Int64Var(&t, "t", 100, "设置时长")
}

type Body struct {
	// Message from user id
	From string `json:"from"`
	// Message to user id
	To string `json:"to"`
	// Message type
	Category string `json:"category"`
	// Message content
	Content string `json:"content"`
	// Message Extend Field
	Ext interface{} `json:"ext"`
	// Unix timestrap to send the message
	SendAt int64 `json:"send_at"`
	ReadAt int64 `json:"read_at"`
	// Unix timestrap to received the message
	ReceivedAt int64 `json:"received_at"`
}

type Payload struct {
	Action string `json:"action"`
	Body   Body   `json:"body"`
	Tries  uint8  `json:"tries"`
}

func main() {
	flag.Parse()
	if help {
		flag.Usage()
	}

	var c int

	var conss = make(map[int]*websocket.Conn, n)

	for c = 1; c <= n; c++ {
		conss[c] = connectionGenerate(c)
		//go messageGenerate(string(c), conn)
	}

	time.Sleep(1000e9)
}

func connectionGenerate(id int) *websocket.Conn {
	var con = websocket.Dialer{
		HandshakeTimeout: 10e9,
	}
	var header = http.Header{}
	url := fmt.Sprintf("ws://127.0.0.1:9501/im?id=%v&version=1.0.1&platform=ios", id)
	conn, res, err := con.Dial(url, header)
	if err != nil {
		log.Println(res)
		log.Printf("%v", err)
	}
	return conn
}

func messageGenerate(id string, conn *websocket.Conn) {
	//end := time.Now().Add(t * time.Second)
	//conn.WriteJSON()
}
