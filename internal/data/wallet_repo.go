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

func (r *WalletRepo) CheckUserExist(userID int64) (bool, error) {
	tx := r.db.First(&domain.User{ID: userID})
	if err := tx.Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return tx.RowsAffected > 0, nil
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
		;
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
		OtherID:        req.TargetID, // properly handle nil and non-nil case for deposit and transfer
		Amount:         req.Amt,
		IsDebit:        false,
		CreateAt:       now,
	}
	t := r.db.Create(&trans)
	if err := t.Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return errs.ErrDuplicate
		}
		return err
	}
	if t.RowsAffected < 1 {
		return errs.ErrNotAllowed
	}
	return nil
}

func (r *WalletRepo) Withdraw(req domain.TransactionReq) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// lock the users row to prevent from race considtion when checking balance
		if err := tx.Exec("select id from users where id = ? for update", req.UserID).Error; err != nil {
			return err
		}

		// for `confirm` phase of transfer, before creating a new transaction record, should update status to confirmed
		if req.Type == domain.Transfer {
			now := time.Now()
			fb := &domain.FrozenBalance{
				IdempotencyKey: req.IdempotencyKey,
			}
			if err := tx.Model(&fb).
				Where("status", domain.FrozenStatusFrozen).
				Updates(map[string]any{
					"status":    domain.FrozenStatusConfirmed,
					"update_at": now,
				}).Error; err != nil {
				return err
			}
		}

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

		INSERT INTO transactions (idempotencykey, userid, otherid, amt, isdebit, at)
		SELECT @idpkey, a.userid, @otherid, @amt, true, NOW()
		FROM available a
		WHERE a.available_balance - @amt >= 0 and @amt > 0
		;
	`
		t := tx.Exec(sql, map[string]any{
			"idpkey":  req.IdempotencyKey,
			"uid":     req.UserID,
			"otherid": req.TargetID, // properly handle nil and non-nil case for deposit and transfer
			"status":  domain.FrozenStatusFrozen,
			"amt":     req.Amt,
		})
		if err := t.Error; err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return errs.ErrDuplicate
			}
			return err
		}
		if t.RowsAffected < 1 {
			return errs.ErrInsufficientBalance
		}
		return nil
	})
}

func (r *WalletRepo) CreateFrozenBalance(req domain.TransactionReq) error {
	rc := domain.FrozenBalance{
		IdempotencyKey: req.IdempotencyKey,
		UserID:         req.UserID,
		TargetID:       *req.TargetID,
		Amount:         req.Amt,
		Status:         domain.FrozenStatusFrozen,
		CreateAt:       time.Now(),
	}
	t := r.db.Create(&rc)
	if err := t.Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return errs.ErrDuplicate
		}
		return err
	}
	if t.RowsAffected < 1 {
		return errs.ErrNotAllowed
	}
	return nil
}

func (r *WalletRepo) CancelFrozenBalance(req domain.TransactionReq) error {
	fb := &domain.FrozenBalance{
		IdempotencyKey: req.IdempotencyKey,
	}
	now := time.Now()
	if err := r.db.Model(&fb).
		Where("status", domain.FrozenStatusFrozen).
		Updates(map[string]any{
			"status":    domain.FrozenStatusCancelled,
			"update_at": now,
		}).Error; err != nil {
		slog.Error("failed to cancel", "req", req, "err", err)
		return err
	}
	return nil
}
