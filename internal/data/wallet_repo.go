package data

import (
	"errors"
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
		FROM users a
		LEFT JOIN (
			SELECT userid, SUM(CASE WHEN isdebit THEN -amt ELSE amt END) AS total_balance
			FROM transactions
			GROUP BY userid
		) t ON t.userid = a.id
		LEFT JOIN (
			SELECT userid, SUM(amt) AS frozen_balance
			FROM frozen_balances
			WHERE status = ?
			GROUP BY userid
		) f ON f.userid = a.id
		WHERE a.id = ?;
	`
	var balanceInfo domain.BalanceBaseInfo
	if err := r.db.Raw(sql, domain.FrozenStatusFrozen, userID).Scan(&balanceInfo).Error; err != nil {
		return nil, err
	}
	return &balanceInfo, nil
}

func (r *WalletRepo) Deposit(req domain.DepositReq) error {
	now := time.Now()
	trans := domain.Transaction{
		IdempotencyKey: req.IdempotencyKey,
		UserID:         req.UserID,
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
