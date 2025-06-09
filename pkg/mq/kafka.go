package mq

import (
	"time"

	"github.com/segmentio/kafka-go"
)

type TopicCategory string

var (
	TopicCategoryTry     TopicCategory = "try"
	TopicCategoryConfirm TopicCategory = "confirm"
	TopicCategoryCancel  TopicCategory = "cancel"
)

type KafkaCfg struct {
	KafkaAddr string `mapstructure:"addr"`
	Topics    map[TopicCategory]string
}

type KafkaSender struct {
	writers map[TopicCategory]*kafka.Writer // topic => writer
}

func NewKafkaSender(cfg KafkaCfg) *KafkaSender {
	s := &KafkaSender{
		writers: make(map[TopicCategory]*kafka.Writer, len(cfg.Topics)),
	}
	for cat, topic := range cfg.Topics {
		w := &kafka.Writer{
			Addr:         kafka.TCP(cfg.KafkaAddr),
			Balancer:     &kafka.Hash{},
			BatchTimeout: 100 * time.Millisecond,
			Topic:        topic,
		}
		s.writers[cat] = w
	}
	return s
}

type KafkaReceiver struct{}
