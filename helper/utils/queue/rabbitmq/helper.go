package rabbitmq

import (
	"encoding/json"

	"pulse/helper/utils/queue/consumer"
)

func GetExchange(metas map[string]string) string {
	var exchange string
	if v, ok := metas[ExchangeKey]; ok {
		exchange = v
	}
	return exchange
}

func Parse(message []byte) (*consumer.Payload, error) {
	p := consumer.Payload{}
	if err := json.Unmarshal(message, &p); err != nil {
		return nil, err
	}
	return &p, nil
}
