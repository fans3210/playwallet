package tests

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"playwallet/internal/apis"
	"playwallet/internal/cfgs"
	"playwallet/internal/domain"
	"playwallet/pkg/consts"

	"github.com/segmentio/kafka-go"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var errNon200Status = fmt.Errorf("unsuccessful status code")

func provisionTestApp(t *testing.T) (string, *gorm.DB, func(t *testing.T)) {
	// setup global logger
	// logger := slog.New(slog.NewTextHandler(io.Discard, nil)) // FIXME:
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
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
	// t.Logf("test cfg used is: %+v\n", cfg)

	primaryDB, err := gorm.Open(postgres.Open(cfg.PG.DSN()), &gorm.Config{
		TranslateError: true,
	})
	if err != nil {
		t.Fatalf("db conn err: %s\n", err)
	}

	testDB := fmt.Sprintf("testdb%d", time.Now().UnixMilli())
	primaryDB.Exec(fmt.Sprintf("create database %s", testDB))
	t.Logf("create test db: %s\n", testDB)

	// WARN: hack, for test app, after tmp db created, connect to tmp db by modifying the cfg
	tmpCfg := cfg
	tmpCfg.PG.DB = testDB
	tmpDB, err := gorm.Open(postgres.Open(tmpCfg.PG.DSN()), &gorm.Config{
		TranslateError: true,
	})
	if err != nil {
		t.Fatalf("tmp test db conn err: %s\n", err)
	}
	dropDB := func(t *testing.T) {
		if err := primaryDB.Exec(fmt.Sprintf("drop database if exists %s with (force)", testDB)).Error; err != nil {
			t.Errorf("failed to drop test db: %s", err)
			return
		}
		t.Logf("dropped db: %s", testDB)
	}
	if err := createTestAccts(tmpDB); err != nil {
		dropDB(t)
		t.Fatalf("failed to prepare test data, err: %s\n", err)
	}

	app, err := apis.NewApp(tmpCfg)
	if err != nil {
		t.Fatalf("failed to create app: %s\n", err)
	}
	kConn, err := createKafkaTestTopics(tmpCfg.Kafka)
	if err != nil {
		t.Fatalf("failed to create kafka test topics, err: %s\n", err)
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

	cleanup := func(t *testing.T) {
		defer kConn.Close()
		topics := make([]string, 0, len(tmpCfg.Kafka.Topics))
		for _, topic := range tmpCfg.Kafka.Topics {
			topics = append(topics, topic)
		}
		// attepmt to delete topic first to make sure no msg exist to make a clean test env
		if err := kConn.DeleteTopics(topics...); err != nil {
			slog.Warn("failed to delete test topics, err:", "err", err)
		} else {
			slog.Warn("kafka test topics deleted successfully")
		}
		if err := app.ShunDown(); err != nil {
			t.Errorf("failed to shut down server: %s\n", err)
			return
		}
		t.Log("server shutdown")

		dropDB(t)
	}

	return fmt.Sprintf("http://localhost:%d", addr.Port), tmpDB, cleanup
}

// create two test accounts with start balance of 0
func createTestAccts(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&domain.User{},
		&domain.FrozenBalance{},
		&domain.Transaction{},
	); err != nil {
		return fmt.Errorf("fail to create table when preparing test data: %w", err)
	}
	// creat test users
	users := []domain.User{{ID: 1, UserName: "testuser1"}, {ID: 2, UserName: "testuser2"}}
	if err := db.Create(users).Error; err != nil {
		return fmt.Errorf("failed to create test users, err: %w", err)
	}
	return nil
}

type actionType string

const (
	deposit  actionType = "deposit"
	withdraw actionType = "withdraw"
	transfer actionType = "transfer"
	freeze   actionType = "freeze"
)

type action struct {
	userID   int64
	targetID int64 // optional
	actType  actionType
	amt      uint64
}

// just use to mock data, no validation checking
// assume there are already two users with id 1 and 2 for testing
func addTestData(t *testing.T, db *gorm.DB, acts ...action) {
	trans := make([]domain.Transaction, 0, len(acts)*2)
	fbs := make([]domain.FrozenBalance, 0, len(acts))
	now := time.Now()
	for i, act := range acts {
		if act.userID <= 0 {
			continue
		}
		switch act.actType {
		case deposit, withdraw:
			trans = append(trans, domain.Transaction{
				IdempotencyKey: fmt.Sprintf("%d:uid:%d:%s", i, act.userID, act.actType),
				UserID:         act.userID,
				Amount:         int64(act.amt),
				IsDebit:        act.actType == withdraw,
				CreateAt:       now,
			})
		case freeze, transfer:
			if act.targetID <= 0 {
				continue
			}
			fb := domain.FrozenBalance{
				IdempotencyKey: fmt.Sprintf("%d:from:%d:to:%d:%s", i, act.userID, act.targetID, act.actType),
				UserID:         act.userID,
				TargetID:       act.targetID,
				Amount:         int64(act.amt),
				Status:         domain.FrozenStatusFrozen,
				CreateAt:       now,
			}
			if act.actType == transfer {
				trans = append(trans, frozenBalancesToTransactions(now, fb)...)
				fb.Status = domain.FrozenStatusConfirmed // after creating transactions, mark the frozen balance record as `confirmed`
			}
			fbs = append(fbs, fb)
		}
	}
	if len(fbs) > 0 {
		if err := db.Create(fbs).Error; err != nil {
			t.Fatal(fmt.Errorf("failed to create test frozen balance data, err: %w", err))
		}
	}
	b, _ := json.Marshal(fbs)
	t.Logf("fbs data: %s\n", b)
	if len(trans) > 0 {
		if err := db.Create(trans).Error; err != nil {
			t.Fatal(fmt.Errorf("failed to create test transactions data, err: %w", err))
		}
	}
	b, _ = json.Marshal(trans)
	t.Logf("trans data: %s\n", b)
}

func makeCheckBalanceReq(endpoint string, uid int64) (int, *domain.BalanceInfo, error) {
	res, err := http.Get(fmt.Sprintf("%s/balance/%d", endpoint, uid))
	if err != nil {
		return 0, nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return res.StatusCode, nil, errNon200Status
	}
	b, err := io.ReadAll(res.Body)
	if err != nil {
		return res.StatusCode, nil, err
	}
	var bInfo domain.BalanceInfo
	if err := json.Unmarshal(b, &bInfo); err != nil {
		return res.StatusCode, nil, err
	}
	return res.StatusCode, &bInfo, nil
}

func makeTransactionReq(endpoint string, req domain.TransactionReq) (int, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return 0, err
	}
	res, err := http.Post(fmt.Sprintf("%s/transaction", endpoint), "application/json", bytes.NewBuffer(body))
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		if res.StatusCode != http.StatusForbidden {
			b, _ := io.ReadAll(res.Body)
			fmt.Printf("make transaction err non 403, reqbody = %s, status: %d\n", b, res.StatusCode)
		}
		return res.StatusCode, errNon200Status
	}
	b, err := io.ReadAll(res.Body)
	if err != nil {
		return res.StatusCode, err
	}
	mRes := make(map[string]any)
	if err := json.Unmarshal(b, &mRes); err != nil {
		return res.StatusCode, err
	}
	if mRes["message"] != consts.DepositSuccessMsg && mRes["message"] != consts.TransferReqSent {
		return res.StatusCode, fmt.Errorf("unexpected response: %+v", mRes)
	}
	return res.StatusCode, nil
}

// each fronzen balance map to 2 transaction records => one debit, one credit
func frozenBalancesToTransactions(at time.Time, fbs ...domain.FrozenBalance) []domain.Transaction {
	ret := make([]domain.Transaction, 0, 2*len(fbs))
	for _, fb := range fbs {
		if fb.Status == domain.FrozenStatusConfirmed {
			continue
		}
		transDebit := domain.Transaction{
			IdempotencyKey: fb.IdempotencyKey,
			UserID:         fb.UserID,
			OtherID:        &fb.TargetID,
			Amount:         fb.Amount,
			IsDebit:        true,
			CreateAt:       at,
		}
		transCredit := domain.Transaction{
			IdempotencyKey: fb.IdempotencyKey,
			UserID:         fb.TargetID,
			OtherID:        nil,
			Amount:         fb.Amount,
			IsDebit:        false,
			CreateAt:       at,
		}
		ret = append(ret, transDebit, transCredit)
	}
	return ret
}

func createKafkaTestTopics(cfg cfgs.KafkaCfg) (*kafka.Conn, error) {
	conn, err := kafka.Dial("tcp", cfg.KafkaAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to conn kafka, %w", err)
	}
	topicCfgs := make([]kafka.TopicConfig, 0, len(cfg.Topics))
	for _, topic := range cfg.Topics {
		topicCfg := kafka.TopicConfig{
			Topic:             topic,
			NumPartitions:     1,
			ReplicationFactor: 1,
		}
		topicCfgs = append(topicCfgs, topicCfg)
	}
	if err := conn.CreateTopics(topicCfgs...); err != nil {
		defer conn.Close()
		return conn, fmt.Errorf("failed to create test topics, err: %w", err)
	}

	return conn, nil
}
