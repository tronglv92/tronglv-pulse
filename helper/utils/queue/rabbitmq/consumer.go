package rabbitmq

import (
	"context"
	"log"

	"pulse/helper/utils/queue/consumer"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/zeromicro/go-zero/core/logx"
	"go.opentelemetry.io/otel"
)

const spanName = "Amqp.Consumer"

type (
	RabbitListener struct {
		listener  *RabbitConnect
		forever   chan bool
		handler   consumer.MessageHandler
		consumers []ConsumerConf
		closed    chan bool
	}
)

func MustNewListener(c Config, handler consumer.MessageHandler) *RabbitListener {
	r := RabbitListener{
		listener:  MustConnect(c),
		consumers: c.Consumers,
		handler:   handler,
		forever:   make(chan bool),
		closed:    make(chan bool),
	}
	go r.listener.Monitor(
		func() {
			r.listenerClose()
		},
		func() {
			r.closed = make(chan bool)
			r.listenerQueues()
		},
	)
	return &r
}

func (q *RabbitListener) listenerQueues() {
	if err := q.listener.openChannel(); err != nil {
		logx.Error(err)
		return
	}
	for _, que := range q.consumers {
		msg, err := q.listener.Channel.Consume(
			que.Name,
			"",
			que.AutoAck,
			que.Exclusive,
			que.NoLocal,
			que.NoWait,
			nil,
		)
		if err != nil {
			log.Fatalf("failed to listener, error: %v", err)
		}

		go func(name string) {
			logx.Infof("Listening on queue: %s", name)
			for {
				select {
				case <-q.closed:
					return
				default:
					for d := range msg {
						payload, e := Parse(d.Body)
						if e != nil {
							logx.Errorf("Error when parsing message: %s, error: %v", string(d.Body), e)
							continue
						}
						if len(payload.GetQueueName()) == 0 {
							payload.QueueName = GetExchange(payload.GetMetaData())
						}
						logx.Infof("Queue %s with trace id [%s]  processing...", payload.GetQueueName(), payload.GetTraceId())

						args := map[string]any{}
						if err = q.handler.Consume(q.spanContext(d.Headers), payload, args); err != nil {
							logx.Errorf("Error on consuming: %s, error: %v", payload.GetString(), err)
							continue
						}
					}
				}
			}
		}(que.Name)
	}
}

func (q *RabbitListener) listenerClose() {
	close(q.closed)
	_ = q.listener.Channel.Close()
	_ = q.listener.Conn.Close()
}

func (q *RabbitListener) Start() {
	q.listenerQueues()
	<-q.forever
}

func (q *RabbitListener) Stop() {
	q.listenerClose()
	close(q.forever)
}

func (q *RabbitListener) spanContext(headers amqp.Table) context.Context {
	tr := otel.Tracer(TracerKey)
	ctx, span := tr.Start(
		ExtractAMQPHeaders(context.Background(), headers),
		spanName,
	)
	span.End()
	return ctx
}
