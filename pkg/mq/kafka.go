package mq

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"
)

type KafkaSender struct {
	writer *kafka.Writer
}

func NewKafkaSender(addr, topic string) (*KafkaSender, error) {
	if topic == "" || addr == "" {
		return nil, fmt.Errorf("invalid param for addr/topic, must not be empty")
	}
	w := kafka.NewWriter(kafka.WriterConfig{
		Brokers:      []string{addr},
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 100 * time.Millisecond,
		Topic:        topic,
	})
	return &KafkaSender{
		writer: w,
	}, nil
}

func (s *KafkaSender) SendMsg(msgs ...kafka.Message) error {
	if err := s.writer.WriteMessages(context.Background(), msgs...); err != nil {
		return fmt.Errorf("kafka write fail, topic: %s, addr: %s, err: %w", s.writer.Topic, s.writer.Addr, err)
	}
	return nil
}

type KafkaHandler func(kafka.Message) error

type KafkaReceiver struct {
	topic  string
	reader *kafka.Reader
}

func NewKafkaReceiver(addr string, topic string, groupID string) (*KafkaReceiver, error) {
	if topic == "" || addr == "" {
		return nil, fmt.Errorf("invalid param for addr/topic/groupid, must not be empty")
	}
	rCfg := kafka.ReaderConfig{
		Brokers:        []string{addr},
		GroupID:        groupID,
		Topic:          topic,
		MinBytes:       1,
		MaxBytes:       10e6,
		MaxWait:        500 * time.Millisecond,
		StartOffset:    kafka.FirstOffset,
		CommitInterval: time.Second,
	}
	r := kafka.NewReader(rCfg)
	return &KafkaReceiver{reader: r, topic: topic}, nil
}

func (r *KafkaReceiver) StartReceive(ctx context.Context, handler KafkaHandler) {
	slog.Info("kafka receiver start receiving from topic", "topic", r.topic)
	go func() {
		for {
			select {
			case <-ctx.Done():
				slog.Warn("kafka reader close due to ctx cancel, topic", "topic", r.topic)
				return
			default:
				msg, err := r.reader.FetchMessage(ctx)
				if err != nil {
					if errors.Is(err, io.EOF) {
						slog.Warn("kafka reader close due to EOF, topic", "topic", r.topic)
						return
					}
					slog.Error("kafka reader read msg err", "err", err, "topic", r.topic)
				}
				if err := handler(msg); err != nil {
					if errors.Is(err, io.EOF) { // if EOF in handler, likely msg is malformed or implemente, not worth retrying
						slog.Error("malformed msg, commiting to avoid retry for topic", "topic", msg.Topic)
						if err := r.reader.CommitMessages(ctx, msg); err != nil {
							slog.Error("failed to commit msg", "err", err, "topic", r.topic)
						}
						return
					}
					slog.Error("reader failed to process kafka msg", "err", err, "topic", r.topic) // TODO: in the future if exceed max retry, move to dead letter queue
					continue
				}
				if err := r.reader.CommitMessages(ctx, msg); err != nil {
					slog.Error("failed to commit msg", "err", err, "topic", r.topic)
				}
			}
		}
	}()
}
