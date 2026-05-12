package kafka

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io"
	"log"
	"os"
	"time"

	"pulse/helper/utils/errors"
	"pulse/helper/utils/queue/consumer"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
	"github.com/zeromicro/go-zero/core/contextx"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/queue"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/core/stat"
	"github.com/zeromicro/go-zero/core/threading"
	"github.com/zeromicro/go-zero/core/timex"
	"github.com/zeromicro/go-zero/core/trace"
	"go.opentelemetry.io/otel"
)

const (
	defaultCommitInterval = time.Second
	defaultMaxWait        = time.Second
	defaultQueueCapacity  = 1000
)

type (
	ConsumeHandle func(ctx context.Context, key, value string) error

	ConsumeErrorHandler func(ctx context.Context, msg kafka.Message, err error)

	kafkaReader interface {
		FetchMessage(ctx context.Context) (kafka.Message, error)
		CommitMessages(ctx context.Context, msgs ...kafka.Message) error
		Close() error
	}

	queueOptions struct {
		commitInterval time.Duration
		queueCapacity  int
		maxWait        time.Duration
		metrics        *stat.Metrics
		errorHandler   ConsumeErrorHandler
	}

	QueueOption func(*queueOptions)

	kafkaQueue struct {
		c                Config
		consumer         kafkaReader
		handler          consumer.MessageHandler
		channel          chan kafka.Message
		producerRoutines *threading.RoutineGroup
		consumerRoutines *threading.RoutineGroup
		commitRunner     *threading.StableRunner[kafka.Message, kafka.Message]
		metrics          *stat.Metrics
		errorHandler     ConsumeErrorHandler
	}

	kafkaQueues struct {
		topic  string
		queues []queue.MessageQueue
		group  *service.ServiceGroup
	}
)

func MustNewListener(c Config, handler consumer.MessageHandler, opts ...QueueOption) queue.MessageQueue {
	q, err := NewQueue(c, handler, opts...)
	if err != nil {
		log.Fatal(err)
	}
	return q
}

func NewQueue(c Config, handler consumer.MessageHandler, opts ...QueueOption) (queue.MessageQueue, error) {
	if err := c.SetUp(); err != nil {
		return nil, err
	}

	var options queueOptions
	for _, opt := range opts {
		opt(&options)
	}
	ensureQueueOptions(c, &options)

	if c.Conns < 1 {
		c.Conns = 1
	}
	q := kafkaQueues{
		topic: c.Topic,
		group: service.NewServiceGroup(),
	}
	for i := 0; i < c.Conns; i++ {
		q.queues = append(q.queues, newKafkaQueue(c, handler, options))
	}
	return q, nil
}

func newKafkaQueue(c Config, handler consumer.MessageHandler, options queueOptions) queue.MessageQueue {
	var offset int64
	if c.Offset == firstOffset {
		offset = kafka.FirstOffset
	} else {
		offset = kafka.LastOffset
	}

	readerConfig := kafka.ReaderConfig{
		Brokers:        c.Brokers,
		GroupID:        c.Group,
		Topic:          c.Topic,
		StartOffset:    offset,
		MinBytes:       c.MinBytes, // 10KB
		MaxBytes:       c.MaxBytes, // 10MB
		MaxWait:        options.maxWait,
		CommitInterval: options.commitInterval,
		QueueCapacity:  options.queueCapacity,
	}
	if len(c.Username) > 0 && len(c.Password) > 0 {
		readerConfig.Dialer = &kafka.Dialer{
			SASLMechanism: plain.Mechanism{
				Username: c.Username,
				Password: c.Password,
			},
		}
	}
	if len(c.CaFile) > 0 {
		caCert, err := os.ReadFile(c.CaFile)
		if err != nil {
			log.Fatal(err)
		}

		caCertPool := x509.NewCertPool()
		ok := caCertPool.AppendCertsFromPEM(caCert)
		if !ok {
			log.Fatal(err)
		}

		readerConfig.Dialer.TLS = &tls.Config{
			RootCAs:            caCertPool,
			InsecureSkipVerify: true,
		}
	}

	q := &kafkaQueue{
		c:                c,
		consumer:         kafka.NewReader(readerConfig),
		handler:          handler,
		channel:          make(chan kafka.Message),
		producerRoutines: threading.NewRoutineGroup(),
		consumerRoutines: threading.NewRoutineGroup(),
		metrics:          options.metrics,
		errorHandler:     options.errorHandler,
	}
	if c.CommitInOrder {
		q.commitRunner = threading.NewStableRunner(func(msg kafka.Message) kafka.Message {
			payload, e := consumer.Parse(msg.Value)
			if e != nil {
				logx.Errorf("Error when parsing message: %s, error: %v", string(msg.Value), e)
				return kafka.Message{}
			}
			payload.QueueName = msg.Topic

			if err := q.consumeOne(context.Background(), payload, map[string]any{Key: msg.Key}); err != nil {
				if q.errorHandler != nil {
					q.errorHandler(context.Background(), msg, err)
				}
			}
			return msg
		})
	}

	return q
}

func (q *kafkaQueue) Start() {
	if q.c.CommitInOrder {
		go q.commitInOrder()

		if err := q.consume(func(msg kafka.Message) {
			if e := q.commitRunner.Push(msg); e != nil {
				logx.Error(e)
			}
		}); err != nil {
			logx.Error(err)
		}
	} else {
		q.startConsumers()
		q.startProducers()
		q.producerRoutines.Wait()
		close(q.channel)
		q.consumerRoutines.Wait()

		logx.Infof("Consumer %s is closed", q.c.Name)
	}
}

func (q *kafkaQueue) Stop() {
	if err := q.consumer.Close(); err != nil {
		logx.Error(err)
	}
	_ = logx.Close()
}

func (q *kafkaQueue) consumeOne(ctx context.Context, payload consumer.MessageContext, args map[string]any) error {
	startTime := timex.Now()
	err := q.handler.Consume(ctx, payload, args)
	q.metrics.Add(stat.Task{
		Duration: timex.Since(startTime),
	})
	return err
}

func (q *kafkaQueue) startConsumers() {
	tracer := otel.Tracer(trace.TraceName)
	for i := 0; i < q.c.Processors; i++ {
		q.consumerRoutines.Run(func() {
			for msg := range q.channel {
				mc := NewMessageCarrier(NewMessage(&msg))
				ctx := otel.GetTextMapPropagator().Extract(context.Background(), mc)
				ctx = contextx.ValueOnlyFrom(ctx)
				traceCtx, kafkaSpan := tracer.Start(ctx, "kafka.process")

				payload, e := consumer.Parse(msg.Value)
				if e != nil {
					logx.Errorf("Error when parsing message: %s, error: %v", string(msg.Value), e)
					continue
				}
				payload.QueueName = msg.Topic

				logx.Infof("Queue %s with trace id [%s]  processing...", payload.GetQueueName(), payload.GetTraceId())
				if err := q.consumeOne(traceCtx, payload, map[string]any{"key": msg.Key}); err != nil {
					kafkaSpan.RecordError(err)
					if q.errorHandler != nil {
						q.errorHandler(traceCtx, msg, err)
					}

					if !q.c.ForceCommit {
						kafkaSpan.End()
						continue
					}
				}

				if err := q.consumer.CommitMessages(traceCtx, msg); err != nil {
					logc.Errorf(traceCtx, "commit failed, error: %v", err)
					kafkaSpan.RecordError(err)
				}
				kafkaSpan.End()
			}
		})
	}
}

func (q *kafkaQueue) startProducers() {
	for i := 0; i < q.c.Consumers; i++ {
		i := i
		q.producerRoutines.Run(func() {
			if err := q.consume(func(msg kafka.Message) {
				q.channel <- msg
			}); err != nil {
				logx.Infof("Consumer %s-%d is closed, error: %q", q.c.Name, i, err.Error())
				return
			}
		})
	}
}

func (q *kafkaQueue) consume(handle func(msg kafka.Message)) error {
	for {
		msg, err := q.consumer.FetchMessage(context.Background())
		// io.EOF means consumer closed
		// io.ErrClosedPipe means committing messages on the consumer,
		// kafka will refire the messages on uncommitted messages, ignore
		if err == io.EOF || errors.Is(err, io.ErrClosedPipe) {
			return err
		}
		if err != nil {
			logx.Errorf("Error on reading message, %q", err.Error())
			continue
		}
		handle(msg)
	}
}

func (q *kafkaQueue) commitInOrder() {
	for {
		msg, err := q.commitRunner.Get()
		if err != nil {
			logx.Error(err)
			return
		}

		if err := q.consumer.CommitMessages(context.Background(), msg); err != nil {
			logx.Errorf("commit failed, error: %v", err)
		}
	}
}

func (q kafkaQueues) Start() {
	for _, each := range q.queues {
		q.group.Add(each)
	}
	logx.Infof("Listening on topic: %s", q.topic)
	q.group.Start()
}

func (q kafkaQueues) Stop() {
	q.group.Stop()
}

func WithCommitInterval(interval time.Duration) QueueOption {
	return func(options *queueOptions) {
		options.commitInterval = interval
	}
}

func WithQueueCapacity(queueCapacity int) QueueOption {
	return func(options *queueOptions) {
		options.queueCapacity = queueCapacity
	}
}

func WithHandle(handle consumer.MessageHandler) consumer.MessageHandler {
	return innerConsumeHandler{
		handle: handle,
	}
}

func WithMaxWait(wait time.Duration) QueueOption {
	return func(options *queueOptions) {
		options.maxWait = wait
	}
}

func WithMetrics(metrics *stat.Metrics) QueueOption {
	return func(options *queueOptions) {
		options.metrics = metrics
	}
}

func WithErrorHandler(errorHandler ConsumeErrorHandler) QueueOption {
	return func(options *queueOptions) {
		options.errorHandler = errorHandler
	}
}

type innerConsumeHandler struct {
	handle consumer.MessageHandler
}

func (ch innerConsumeHandler) Consume(ctx context.Context, payload consumer.MessageContext, args map[string]any) error {
	return ch.handle.Consume(ctx, payload, args)
}

func ensureQueueOptions(c Config, options *queueOptions) {
	if options.commitInterval == 0 {
		options.commitInterval = defaultCommitInterval
	}
	if options.queueCapacity == 0 {
		options.queueCapacity = defaultQueueCapacity
	}
	if options.maxWait == 0 {
		options.maxWait = defaultMaxWait
	}
	if options.metrics == nil {
		options.metrics = stat.NewMetrics(c.Name)
	}
	if options.errorHandler == nil {
		options.errorHandler = func(ctx context.Context, msg kafka.Message, err error) {
			logc.Errorf(ctx, "consume: %s, error: %v", string(msg.Value), err)
		}
	}
}
