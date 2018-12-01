package im

// Define message body
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

type Message interface {
	GetAction() string
	GetBody() string
	GetContent() string
	GetReceivedAt() int64
}

type Payload struct {
	Action string `json:"action"`
	Body Body `json:"body"`
	Tries uint8 `json:"tries"`
}


func (payload *Payload) GetAction() string {
	return payload.Action
}

func (payload *Payload) GetBody() Body {
	return payload.Body
}

func (payload *Payload) GetTries() uint8 {
	return payload.Tries
}

func (payload *Payload) GetFrom() string {
	return payload.Body.From
}

func (payload *Payload) GetTo() string {
	return payload.Body.To
}

func (payload *Payload) GetCategory() string {
	return payload.Body.Category
}

func (payload *Payload) GetContent() string {
	return payload.Body.Content
}

func (payload *Payload) GetExt() interface{} {
	return payload.Body.Ext
}

func (payload *Payload) GetSendAt() int64 {
	return payload.Body.SendAt
}

func (payload *Payload) GetReadAt() int64 {
	return payload.Body.ReadAt
}

func (payload *Payload) GetReceivedAt() int64 {
	return payload.Body.ReceivedAt
}