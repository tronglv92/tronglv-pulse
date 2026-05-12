package kafka

import (
	"context"
	"strconv"
	"time"

	"pulse/helper/utils/queue/consumer"
	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
)

type (
	PushOption func(options *pushOptions)

	Producer struct {
		topic    string
		producer kafkaWriter
	}

	kafkaWriter interface {
		Close() error
		WriteMessages(ctx context.Context, msgs ...kafka.Message) error
	}

	pushOptions struct {
		allowAutoTopicCreation bool
		async                  bool
		balancer               kafka.Balancer
		batchSize              int
		batchTimeout           time.Duration
		requiredAck            kafka.RequiredAcks
	}
)

func NewProducer(c Config, opts ...PushOption) *Producer {
	producer := &kafka.Writer{
		Addr:         kafka.TCP(c.Brokers...),
		Balancer:     &kafka.LeastBytes{},
		Compression:  kafka.Snappy,
		BatchSize:    1,                     // small batch
		BatchTimeout: 10 * time.Millisecond, // short wait
		RequiredAcks: kafka.RequireOne,
	}

	var options pushOptions
	for _, opt := range opts {
		opt(&options)
	}

	producer.AllowAutoTopicCreation = options.allowAutoTopicCreation
	if options.balancer != nil {
		producer.Balancer = options.balancer
	}

	if options.batchSize > 0 {
		producer.BatchSize = options.batchSize
	}

	if options.batchTimeout > 0 {
		producer.BatchTimeout = options.batchTimeout
	}

	if options.requiredAck != 0 {
		producer.RequiredAcks = options.requiredAck
	}

	if options.async {
		producer.Async = options.async
	}
	return &Producer{
		producer: producer,
		topic:    c.Topic,
	}
}

func (r *Producer) Close() error {
	return r.producer.Close()
}

func (r *Producer) Name() string {
	return r.topic
}

func (r *Producer) Send(payload consumer.MessageContext) error {
	return r.SendCtx(context.Background(), payload)
}

func (r *Producer) SendCtx(ctx context.Context, p consumer.MessageContext) error {
	var key = strconv.FormatInt(time.Now().UnixNano(), 10)
	if v, ok := p.GetMetaData()[Key]; ok && len(v) > 0 {
		key = v
	}
	return r.PushWithKey(ctx, key, p)
}

func (r *Producer) PushWithKey(ctx context.Context, key string, p consumer.MessageContext) error {
	msg := kafka.Message{
		Key:   []byte(key),
		Value: []byte(p.GetString()),
		Topic: p.GetQueueName(),
	}

	mc := NewMessageCarrier(NewMessage(&msg))
	otel.GetTextMapPropagator().Inject(ctx, mc)
	return r.producer.WriteMessages(ctx, msg)
}

func WithAllowAutoTopicCreation() PushOption {
	return func(options *pushOptions) {
		options.allowAutoTopicCreation = true
	}
}

func WithBalancer(balancer kafka.Balancer) PushOption {
	return func(options *pushOptions) {
		options.balancer = balancer
	}
}

func WithBatchTimeout(timeout time.Duration) PushOption {
	return func(options *pushOptions) {
		options.batchTimeout = timeout
	}
}

func WithBatchSize(batchSize int) PushOption {
	return func(options *pushOptions) {
		if batchSize <= 0 {
			batchSize = 1
		}
		options.batchSize = batchSize
	}
}

func WithRequiredAck(requiredAck kafka.RequiredAcks) PushOption {
	return func(options *pushOptions) {
		options.requiredAck = requiredAck
	}
}

func WithAsync(async bool) PushOption {
	return func(options *pushOptions) {
		options.async = async
	}
}
