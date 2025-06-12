package biz

import (
	"fmt"

	"playwallet/internal/cfgs"
	"playwallet/internal/data"
	"playwallet/pkg/mq"
)

type WalletUC struct {
	cfg     cfgs.Config
	repo    *data.WalletRepo
	senders map[cfgs.TopicCategory]*mq.KafkaSender // topic => writer
}

func NewWalletUC(cfg cfgs.Config, repo *data.WalletRepo) (*WalletUC, error) {
	uc := &WalletUC{
		cfg:     cfg,
		repo:    repo,
		senders: make(map[cfgs.TopicCategory]*mq.KafkaSender, len(cfg.Kafka.Topics)),
	}
	for cat, topic := range cfg.Kafka.Topics {
		sender, err := mq.NewKafkaSender(cfg.Kafka.KafkaAddr, topic)
		if err != nil {
			return nil, fmt.Errorf("failed to create wallet uc for topiccat: %s, %w", cat, err)
		}
		uc.senders[cat] = sender
	}
	return uc, nil
}
