package rabbitmq

import (
	"context"

	"pulse/helper/utils/queue/consumer"
)

const ExchangeKey = "exchange"

type Producer struct {
	sender *RabbitConnect
}

func NewProducer(cfg Config) *Producer {
	p := &Producer{
		sender: MustConnect(cfg),
	}
	go p.sender.Monitor(nil, nil)
	return p
}

func (r *Producer) Send(payload consumer.MessageContext) error {
	return r.SendCtx(context.Background(), payload)
}

func (r *Producer) SendCtx(ctx context.Context, p consumer.MessageContext) error {
	if err := r.sender.openChannel(); err != nil {
		return err
	}
	defer r.sender.closeChannel()

	err := r.sender.Send(ctx, GetExchange(p.GetMetaData()), p.GetQueueName(), p.GetByte())
	if err != nil {
		return err
	}
	return nil
}

func (r *Producer) Close() error {
	return r.sender.Close()
}
