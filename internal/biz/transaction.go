package biz

import (
	"fmt"

	"playwallet/internal/domain"
	"playwallet/pkg/errs"
)

func (uc *WalletUC) MakeTransaction(req domain.TransactionReq) error {
	exists, err := uc.repo.CheckUserExist(req.UserID)
	if err != nil {
		return fmt.Errorf("make transaction, check user %d exist err, %w", req.UserID, err)
	}
	if !exists {
		return fmt.Errorf("make transaction, user %d not exist, %w", req.UserID, errs.ErrNotFound)
	}
	if req.TargetID != nil && req.Type == domain.Transfer {
		exists, err := uc.repo.CheckUserExist(*req.TargetID)
		if err != nil {
			return fmt.Errorf("make transaction, check target %d exist err, %w", *req.TargetID, err)
		}
		if !exists {
			return fmt.Errorf("make transaction, target %d not exist, %w", *req.TargetID, errs.ErrNotFound)
		}
	}
	switch req.Type {
	case domain.Deposit:
		return uc.repo.Deposit(req)
	case domain.Withdraw:
		return uc.repo.Withdraw(req)
	case domain.Transfer:
		return uc.tccTry(req)
	}
	return fmt.Errorf("not handled yet")
}

func (uc *WalletUC) Transactions(userID int64, pageOpt domain.PageOpt) (int64, []domain.Transaction, error) {
	exists, err := uc.repo.CheckUserExist(userID)
	if err != nil {
		return 0, nil, fmt.Errorf("get transactions, check user %d exist err, %w", userID, err)
	}
	if !exists {
		return 0, nil, fmt.Errorf("get transactions, user %d not exist, %w", userID, errs.ErrNotFound)
	}
	return uc.repo.Transactions(userID, pageOpt)
}
