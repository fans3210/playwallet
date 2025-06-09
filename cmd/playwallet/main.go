package main

import (
	"log/slog"
	"os"

	"playwallet/internal/apis"
	"playwallet/internal/cfgs"
	"playwallet/internal/data"
	"playwallet/pkg/mq"

	"github.com/spf13/viper"
)

func main() {
	// read config
	viper.SetConfigName("cfg")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("config")
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
	cfg := &cfgs.Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		panic(err)
	}

	// setup global logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.Level(cfg.LogLv),
	}))
	slog.SetDefault(logger)

	// db
	repo, err := data.NewPGRepo(cfg.PG)
	if err != nil {
		panic(err)
	}
	_ = repo
	slog.Info("connect to postgres successfully\n")

	// kafk
	kafkaSender := mq.NewKafkaSender(cfg.Kafka)
	slog.Info("init kafka sender successfully\n")
	_ = kafkaSender

	svr, err := apis.NewServer()
	if err != nil {
		panic(err)
	}
	svr.Start()
}
