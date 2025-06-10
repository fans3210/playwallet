package tests

import (
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"playwallet/internal/apis"
	"playwallet/internal/cfgs"

	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func provisionTestApp(t *testing.T) (string, func(t *testing.T)) {
	// setup global logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)
	// read config
	viper.SetConfigName("cfg_test")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("../config")
	if err := viper.ReadInConfig(); err != nil {
		t.Fatalf("failed to read cfg: %s\n", err)
	}
	cfg := cfgs.Config{}
	if err := viper.Unmarshal(&cfg); err != nil {
		t.Fatalf("failed to unmarshal cfg: %s\n", err)
	}
	t.Logf("test cfg used is: %+v\n", cfg)

	db, err := gorm.Open(postgres.Open(cfg.PG.DSN()), &gorm.Config{})
	if err != nil {
		t.Fatalf("db conn err: %s\n", err)
	}

	testDB := fmt.Sprintf("testdb%d", time.Now().UnixMilli())
	db.Exec(fmt.Sprintf("create database %s", testDB))
	t.Logf("create test db: %s\n", testDB)

	// WARN: hack, for test app, after tmp db created, connect to tmp db by modifying the cfg
	cfg.PG.DB = testDB
	app, err := apis.NewApp(cfg)
	if err != nil {
		t.Fatalf("failed to create app: %s\n", err)
	}
	ln, err := app.NewListener()
	if err != nil {
		t.Fatalf("failed to create listener: %s\n", err)
	}
	go func() {
		// random port
		if err := app.StartWithListener(ln); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				return
			}
			t.Errorf("server start err: %s\n", err)
		}
	}()

	addr, ok := ln.Addr().(*net.TCPAddr)
	if !ok {
		t.Fatalf("unable to get tcp addr: %s\n", addr)
	}

	return fmt.Sprintf("http://localhost:%d", addr.Port), func(t *testing.T) {
		if err := app.ShunDown(); err != nil {
			t.Errorf("failed to shut down server: %s\n", err)
			return
		}
		t.Log("server shutdown")

		if err := db.Exec(fmt.Sprintf("drop database if exists %s with (force)", testDB)).Error; err != nil {
			t.Errorf("failed to drop test db: %s", err)
			return
		}
		t.Logf("dropped db: %s", testDB)
	}
}
