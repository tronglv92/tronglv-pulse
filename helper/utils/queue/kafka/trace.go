package kafka

import "go.opentelemetry.io/otel/propagation"

var _ propagation.TextMapCarrier = (*MessageCarrier)(nil)

type MessageCarrier struct {
	msg *Message
}

func NewMessageCarrier(msg *Message) MessageCarrier {
	return MessageCarrier{msg: msg}
}

func (m MessageCarrier) Get(key string) string {
	return m.msg.GetHeader(key)
}

func (m MessageCarrier) Set(key string, value string) {
	m.msg.SetHeader(key, value)
}

func (m MessageCarrier) Keys() []string {
	out := make([]string, len(m.msg.Headers))
	for i, h := range m.msg.Headers {
		out[i] = h.Key
	}

	return out
}
