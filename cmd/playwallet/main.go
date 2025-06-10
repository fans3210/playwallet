package main

import (
	"log"
	"log/slog"
	"os"

	"playwallet/internal/apis"
	"playwallet/internal/cfgs"

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
	cfg := cfgs.Config{}
	if err := viper.Unmarshal(&cfg); err != nil {
		panic(err)
	}
	// setup global logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.Level(cfg.LogLv),
	}))
	slog.SetDefault(logger)

	svr, err := apis.NewApp(cfg)
	if err != nil {
		panic(err)
	}
	log.Fatal(svr.Start())
}
