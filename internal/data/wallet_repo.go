package data

import (
	"errors"
	"log/slog"
	"time"

	"playwallet/internal/cfgs"
	"playwallet/internal/domain"
	"playwallet/pkg/errs"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	// "gorm.io/gorm/logger"
)

type WalletRepo struct {
	db *gorm.DB
}

func NewWalletRepo(cfg cfgs.PGCfg) (*WalletRepo, error) {
	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		// Logger: logger.Default.LogMode(logger.Silent),
		TranslateError: true,
	})
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(
		&domain.User{},
		&domain.FrozenBalance{},
		&domain.Transaction{},
	); err != nil {
		return nil, err
	}
	ret := &WalletRepo{db}
	return ret, nil
}

func (r *WalletRepo) CheckUserExist(userID int64) bool {
	return r.db.First(&domain.User{ID: userID}).RowsAffected > 0
}

// TODO: instead of scanning all transactions, add a milestone or snapshot table to save the balance before certain date for performance optimisations
func (r *WalletRepo) CheckBalance(userID int64) (*domain.BalanceBaseInfo, error) {
	sql := `
		SELECT 
		  a.id,
		  COALESCE(t.total_balance, 0) AS total_balance,
		  COALESCE(f.frozen_balance, 0) AS frozen_balance
		FROM (select id from users where id = @uid) a
		LEFT JOIN (
			SELECT userid, SUM(CASE WHEN isdebit THEN -amt ELSE amt END) AS total_balance
			FROM transactions
			where userid = @uid
			GROUP BY userid
		) t ON t.userid = a.id
		LEFT JOIN (
			SELECT userid, SUM(amt) AS frozen_balance
			FROM frozen_balances
			WHERE userid = @uid and status = @status
			GROUP BY userid
		) f ON f.userid = a.id
	`
	var balanceInfo domain.BalanceBaseInfo
	if err := r.db.Raw(sql, map[string]any{"uid": userID, "status": domain.FrozenStatusFrozen}).Scan(&balanceInfo).Error; err != nil {
		return nil, err
	}
	return &balanceInfo, nil
}

func (r *WalletRepo) Deposit(req domain.TransactionReq) error {
	now := time.Now()
	trans := domain.Transaction{
		IdempotencyKey: req.IdempotencyKey,
		UserID:         req.UserID,
		TargetID:       req.TargetID, // properly handle nil and non-nil case for deposit and transfer
		Amount:         req.Amt,
		IsDebit:        false,
		At:             now,
	}
	if err := r.db.Create(&trans).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return errs.ErrDuplicate
		}
		return err
	}
	return nil
}

func (r *WalletRepo) Withdraw(req domain.TransactionReq) error {
	sql := `
		WITH available AS (
		  SELECT 
			a.id AS userid,
			COALESCE(t.total_balance, 0) - COALESCE(f.frozen_balance, 0) AS available_balance
			FROM (select id from users where id = @uid) a
		  LEFT JOIN (
			SELECT userid, SUM(CASE WHEN isdebit THEN -amt ELSE amt END) AS total_balance
			FROM transactions
			where userid = @uid
			GROUP BY userid
		  ) t ON t.userid = a.id
		  LEFT JOIN (
			SELECT userid, SUM(amt) AS frozen_balance
			FROM frozen_balances
			WHERE userid = @uid and status = @status
			GROUP BY userid
		  ) f ON f.userid = a.id
		)

		INSERT INTO transactions (id, userid, targetid, amt, isdebit, at)
		SELECT @idpkey, a.userid, @targetid, @amt, true, NOW()
		FROM available a
		WHERE a.available_balance - @amt >= 0 and @amt > 0;
	`
	res := map[string]any{}
	if err := r.db.Debug().Raw(sql, map[string]any{
		"uid":      req.UserID,
		"targetid": req.TargetID, // properly handle nil and non-nil case for deposit and transfer
		"status":   domain.FrozenStatusFrozen,
		"amt":      req.Amt,
	}).Scan(&res).Error; err != nil {
		return err
	}
	slog.Debug("withdraw res = ", "res", res)
	panic("hi")
}
