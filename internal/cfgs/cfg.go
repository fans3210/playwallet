package cfgs

import (
	"fmt"
)

type Config struct {
	Env   string   `mapstructure:"env"`
	LogLv int      `mapstructure:"loglv"`
	Kafka KafkaCfg `mapstructure:"kafka"`
	PG    PGCfg    `mapstructure:"pg"`
	Http  HttpCfg  `mapstructure:"http"`
}

type PGCfg struct {
	Addr     string `mapstructure:"addr"`
	Pwd      string `mapstructure:"pwd"`
	UserName string `mapstructure:"user"`
	DB       string `mapstructure:"db"`
	Port     int    `mapstructure:"port"`
}

func (p PGCfg) DSN() string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Singapore",
		p.Addr, p.UserName, p.Pwd, p.DB, p.Port)
}

type TopicKey string

var (
	TpcKeySenderConfirm   TopicKey = "senderconfirm"
	TpcKeyReceiverConfirm TopicKey = "receiverconfirm"
	TpcKeyCancel          TopicKey = "cancel"
)

type KafkaCfg struct {
	KafkaAddr     string `mapstructure:"addr"`
	Topics        map[TopicKey]string
	ConsumerGroup string `mapstructure:"consumer_group"`
}

type HttpCfg struct {
	Addr string `mapstructure:"addr"`
}
