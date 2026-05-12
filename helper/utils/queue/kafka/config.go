package kafka

import "github.com/zeromicro/go-zero/core/service"

const (
	Key         = "key"
	firstOffset = "first"
	lastOffset  = "last"
)

type Config struct {
	service.ServiceConf
	Brokers       []string `json:"brokers"`
	Group         string   `json:"group"`
	Topic         string   `json:"topic,default=local"`
	Topics        []string `json:"topics"`
	CaFile        string   `json:"ca-file,optional"`
	Offset        string   `json:"offset,options=first|last,default=last"`
	Conns         int      `json:"conns,default=1"`
	Consumers     int      `json:"consumers,default=8"`
	Processors    int      `json:"processors,default=8"`
	MinBytes      int      `json:"min-bytes,default=10240"`    // 10K
	MaxBytes      int      `json:"max-bytes,default=10485760"` // 10M
	Username      string   `json:"username,optional"`
	Password      string   `json:"password,optional"`
	ForceCommit   bool     `json:"force-commit,default=true"`
	CommitInOrder bool     `json:"commit-in-order,default=false"`
}

func (c Config) GetKafkaWithTopic(topic string) Config {
	c.Topic = topic
	return c
}

func (c Config) GetKafkaTopics() []string {
	return c.Topics
}
