package rabbitmq

type (
	callBack func()
)

type Config struct {
	Name       string         `json:"name,optional"`
	Connection RabbitConf     `json:"connection,optional"`
	Consumers  []ConsumerConf `json:"consumers,optional"`
	Queues     []QueueConf    `json:"queues,optional"`
	Exchanges  []ExchangeConf `json:"exchanges,optional"`
}

type RabbitConf struct {
	Username string
	Password string
	Host     string
	Port     int
	VHost    string `json:",optional"`
}

type ConsumerConf struct {
	Name      string `json:"name"`
	AutoAck   bool   `json:"auto-ack,default=true"`
	Exclusive bool   `json:"exclusive,default=false"`
	// Set to true, which means that messages sent by producers in the same connection
	// cannot be delivered to consumers in this connection.
	NoLocal bool `json:"no-local,default=false"`
	// Whether to block processing
	NoWait bool `json:"no-wait,default=false"`
}

type ExchangeConf struct {
	Name       string `json:"name"`
	Type       string `json:"type,default=direct,options=direct|fanout|topic|headers"` // exchange type
	Durable    bool   `json:"durable,default=true"`
	AutoDelete bool   `json:"auto-delete,default=false"`
	Internal   bool   `json:"internal,default=false"`
	NoWait     bool   `json:"no-wait,default=false"`
}

type QueueConf struct {
	Name       string      `json:"name"`
	Durable    bool        `json:"durable,default=true"`
	AutoDelete bool        `json:"auto-delete,default=false"`
	Exclusive  bool        `json:"exclusive,default=false"`
	NoWait     bool        `json:"no-wait,default=false"`
	Binds      []QueueBind `json:"binds,optional"`
}

type QueueBind struct {
	RoutingKey string `json:"routing-key,optional"`
	Exchange   string `json:"exchange,optional"`
	NoWait     bool   `json:"no-wait,optional"`
}
