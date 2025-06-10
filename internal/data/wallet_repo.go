package data

import (
	"log/slog"

	"playwallet/internal/cfgs"
	"playwallet/internal/domain"

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

func (r *WalletRepo) CheckBalance(userID int64) (int64, error) {
	sql := `
	SELECT 
	  SUM(t.amt) AS total_balance,
	  SUM(f.amt) AS frozen_balance
	FROM users u
	LEFT JOIN transactions t ON t.userid = u.id
	LEFT JOIN frozen_balances f ON f.userid = u.id AND f.status = ?
	WHERE u.id = ?
	GROUP BY u.id;
	`
	var balanceInfo domain.BalanceInfo
	if err := r.db.Raw(sql, domain.FrozenStatusFrozen, userID).Scan(&balanceInfo).Error; err != nil {
		return 0, err
	}
	slog.Debug("balance is", "balance", balanceInfo)
	balance := max(balanceInfo.TotalBalance-balanceInfo.FrozenBalance, 0)
	return balance, nil
}
