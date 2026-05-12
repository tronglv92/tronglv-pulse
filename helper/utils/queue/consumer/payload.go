package consumer

import (
	"encoding/json"

	"github.com/google/uuid"
)

type MessageContext interface {
	GetTraceId() string
	GetQueueName() string
	GetMessage() any
	GetByte() []byte
	GetString() string
	GetMetaData() map[string]string
}

type Payload struct {
	TraceId   string            `json:"trace_id"`
	QueueName string            `json:"queue"`
	Message   any               `json:"message"`
	MetaData  map[string]string `json:"meta_data"`
}

func (p *Payload) GetTraceId() string {
	return p.TraceId
}
func (p *Payload) GetQueueName() string {
	return p.QueueName
}
func (p *Payload) GetMessage() any {
	return p.Message
}
func (p *Payload) GetMetaData() map[string]string {
	return p.MetaData
}
func (p *Payload) Values() Payload {
	if len(p.TraceId) == 0 {
		p.TraceId = uuid.NewString()
	}
	return *p
}
func (p *Payload) GetByte() []byte {
	msg, err := json.Marshal(p.Values())
	if err != nil {
		return nil
	}
	return msg
}
func (p *Payload) GetString() string {
	return string(p.GetByte())
}

func Parse(message []byte) (*Payload, error) {
	p := Payload{}
	if err := json.Unmarshal(message, &p); err != nil {
		return nil, err
	}
	return &p, nil
}
