package rabbitmq

import (
	"context"
	"pulse/helper/utils/toolkit/timex"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/zeromicro/go-zero/core/logx"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	oteltrace "go.opentelemetry.io/otel/trace"
	"log"
	"time"
)

const (
	TracerKey  = "amqp"
	TraceIdKey = "trace_id"
	SpanIdKey  = "span_id"
)

type RabbitConnect struct {
	conf        Config
	Conn        *amqp.Connection
	Channel     *amqp.Channel
	CloseErrs   chan *amqp.Error
	ContentType string
	DelayNumber int
}

func getRabbitURL(rabbitConf RabbitConf) string {
	return fmt.Sprintf("amqp://%s:%s@%s:%d/%s", rabbitConf.Username, rabbitConf.Password,
		rabbitConf.Host, rabbitConf.Port, rabbitConf.VHost)
}

func MustConnect(conf Config) *RabbitConnect {
	r := RabbitConnect{
		conf:        conf,
		DelayNumber: 5,
		ContentType: "application/json",
	}

	if err := r.connect(); err != nil {
		log.Fatalf("rabbitmq: %v", err)
	}

	if err := r.declare(); err != nil {
		log.Fatalf("rabbitmq declare: %v", err)
	}
	return &r
}

func (q *RabbitConnect) connect() error {
	conn, err := amqp.Dial(getRabbitURL(q.conf.Connection))
	if err != nil {
		return fmt.Errorf("failed to connect rabbitmq, error: %v", err)
	}
	q.Conn = conn
	q.CloseErrs = make(chan *amqp.Error)
	q.Conn.NotifyClose(q.CloseErrs)
	return nil
}

func (q *RabbitConnect) reConnect(reConFnc func()) {
	for {
		time.Sleep(time.Duration(q.DelayNumber) * time.Second)
		if err := q.connect(); err == nil {
			if reConFnc != nil {
				reConFnc()
			}
			fmt.Printf("[%s] Reconnected to RabbitMq \n", timex.Now().Format(time.DateTime))
			return
		}
		fmt.Printf("[%s] Reconnect failed, retrying... \n", timex.Now().Format(time.DateTime))
	}
}

func (q *RabbitConnect) isClosed(err *amqp.Error) bool {
	return err == nil || err.Code == amqp.ChannelError
}

func (q *RabbitConnect) openChannel() error {
	channel, err := q.Conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open a channel: %v", err)
	}
	q.Channel = channel
	return nil
}

func (q *RabbitConnect) closeChannel() {
	if err := q.Channel.Close(); err != nil {
		logx.Error(err)
	}
}

func (q *RabbitConnect) Monitor(closeFnc callBack, reConFnc callBack) {
	for {
		closeErr := <-q.CloseErrs
		if !q.isClosed(closeErr) {
			fmt.Printf("[%s] Connection or Channel closed: %v\n", timex.Now().Format(time.DateTime), closeErr)
			if closeFnc != nil {
				closeFnc()
			}
			q.reConnect(reConFnc)
		}
	}
}

func (q *RabbitConnect) Send(ctx context.Context, exchange string, routeKey string, msg []byte) error {
	tr := otel.Tracer(TracerKey)
	amqpCtx, span := tr.Start(ctx, "Amqp.Producer")
	defer span.End()
	span.AddEvent("publisher", oteltrace.WithAttributes(
		attribute.String("publisher.exchange", exchange),
		attribute.String("publisher.routeKey", routeKey),
		attribute.String("publisher.payload", string(msg)),
	))

	err := q.Channel.PublishWithContext(
		ctx,
		exchange,
		routeKey,
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  q.ContentType,
			Body:         msg,
			Headers:      InjectAMQPHeaders(amqpCtx),
		},
	)
	if err != nil {
		span.RecordError(err)
	}
	return err
}

func (q *RabbitConnect) Close() error {
	if err := q.Conn.Close(); err != nil {
		return err
	}
	return nil
}

func (q *RabbitConnect) declare() error {
	if err := q.openChannel(); err != nil {
		return err
	}
	defer q.closeChannel()

	if len(q.conf.Exchanges) > 0 {
		if err := q.declareExchange(q.conf.Exchanges); err != nil {
			return err
		}
	}

	if len(q.conf.Queues) > 0 {
		if err := q.declareQueue(q.conf.Queues); err != nil {
			return err
		}
	}
	return nil
}

func (q *RabbitConnect) declareExchange(exchanges []ExchangeConf) error {
	for _, val := range exchanges {
		err := q.Channel.ExchangeDeclare(
			val.Name,
			val.Type,
			val.Durable,
			val.AutoDelete,
			val.Internal,
			val.NoWait,
			nil,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (q *RabbitConnect) declareQueue(queues []QueueConf) error {
	for _, val := range queues {
		_, err := q.Channel.QueueDeclare(
			val.Name,
			val.Durable,
			val.AutoDelete,
			val.Exclusive,
			val.NoWait,
			nil,
		)
		if err != nil {
			return err
		}

		if len(val.Binds) > 0 {
			for _, v := range val.Binds {
				return q.queueBind(val.Name, v)
			}
		}
	}
	return nil
}

func (q *RabbitConnect) queueBind(name string, bind QueueBind) error {
	if err := q.Channel.QueueBind(name, bind.RoutingKey, bind.Exchange, bind.NoWait, nil); err != nil {
		return err
	}
	return nil
}
