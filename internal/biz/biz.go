package biz

import (
	"context"
	"fmt"

	"playwallet/internal/cfgs"
	"playwallet/internal/data"
	"playwallet/pkg/mq"
)

type WalletUC struct {
	cfg     cfgs.Config
	repo    *data.WalletRepo
	senders map[cfgs.TopicKey]*mq.KafkaSender // topic => writer
}

func NewWalletUC(ctx context.Context, cfg cfgs.Config, repo *data.WalletRepo) (*WalletUC, error) {
	uc := &WalletUC{
		cfg:     cfg,
		repo:    repo,
		senders: make(map[cfgs.TopicKey]*mq.KafkaSender, len(cfg.Kafka.Topics)),
	}
	for k, topic := range cfg.Kafka.Topics {
		sender, err := mq.NewKafkaSender(cfg.Kafka.KafkaAddr, topic)
		if err != nil {
			return nil, fmt.Errorf("failed to create kafka sender for topic key: %s, %w", k, err)
		}
		uc.senders[k] = sender

		krecv, err := mq.NewKafkaReceiver(cfg.Kafka.KafkaAddr, topic, cfg.Kafka.ConsumerGroup)
		if err != nil {
			return nil, fmt.Errorf("failed to create kafka receiver for topic key: %s, %w", k, err)
		}
		switch k {
		case cfgs.TpcKeySenderConfirm:
			krecv.StartReceive(ctx, uc.handleSenderConfirm)
		case cfgs.TpcKeyReceiverConfirm:
			krecv.StartReceive(ctx, uc.handleReceiverConfirm)
		case cfgs.TpcKeyCancel:
			krecv.StartReceive(ctx, uc.handleCancel)
		default:
			return nil, fmt.Errorf("unexpected kafka topic key: %s", k)
		}
	}
	return uc, nil
}
