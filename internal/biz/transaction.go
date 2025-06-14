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
