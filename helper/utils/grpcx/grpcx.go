package grpcx

import (
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"sync"
	"time"
)

type ClientBuilder[C any] func(conn *grpc.ClientConn) C

type ClientManager[C any] struct {
	config   zrpc.RpcClientConf
	conn     *grpc.ClientConn
	client   C
	build    ClientBuilder[C]
	mu       sync.Mutex
	interval time.Duration
	stopCh   chan struct{}

	startOnce sync.Once
	started   bool
}

func NewClientManager[C any](conf zrpc.RpcClientConf, build ClientBuilder[C]) *ClientManager[C] {
	return &ClientManager[C]{
		config: conf,
		build:  build,
		stopCh: make(chan struct{}),
	}
}
func (m *ClientManager[C]) Get() (C, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.started {
		m.StartMonitor(30 * time.Second)
	}

	if m.conn != nil && m.conn.GetState() != connectivity.Shutdown {
		return m.client, nil
	}

	if m.conn != nil {
		_ = m.conn.Close()
	}

	client, err := zrpc.NewClient(m.config)
	if err != nil {
		return m.client, err
	}
	m.conn = client.Conn()
	m.client = m.build(m.conn)

	return m.client, nil
}

func (m *ClientManager[C]) StartMonitor(interval time.Duration) {
	m.startOnce.Do(func() {
		m.interval = interval
		m.stopCh = make(chan struct{})
		m.started = true
		go func() {
			ticker := time.NewTicker(interval)
			for {
				select {
				case <-ticker.C:
					m.mu.Lock()
					if m.conn == nil || m.conn.GetState() == connectivity.TransientFailure || m.conn.GetState() == connectivity.Shutdown {
						if m.conn != nil {
							_ = m.conn.Close()
						}
						client, err := zrpc.NewClient(m.config)
						if err != nil {
							logx.Info("failed to reconnect gRPC client: %v", err)
							m.mu.Unlock()
							continue
						}
						m.conn = client.Conn()
						m.client = m.build(m.conn)
					}
					m.mu.Unlock()
				case <-m.stopCh:
					ticker.Stop()
					return
				}
			}
		}()
	})
}

func (m *ClientManager[C]) StopMonitor() {
	close(m.stopCh)
}
