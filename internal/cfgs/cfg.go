package cfgs

import (
	"fmt"

	"playwallet/pkg/mq"
)

type Config struct {
	LogLv int         `mapstructure:"loglv"`
	Kafka mq.KafkaCfg `mapstructure:"kafka"`
	PG    PGCfg       `mapstructure:"pg"`
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
