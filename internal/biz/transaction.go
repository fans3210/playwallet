package biz

import (
	"fmt"

	"playwallet/internal/domain"
	"playwallet/pkg/errs"
)

func (uc *WalletUC) MakeTransaction(req domain.TransactionReq) error {
	if !uc.repo.CheckUserExist(req.UserID) {
		return fmt.Errorf("user not exist, %w", errs.ErrNotFound)
	}
	switch req.Type {
	case domain.Deposit:
		return uc.repo.Deposit(req)
	case domain.Withdraw:
		return uc.repo.Withdraw(req)
	}
	panic("not handled yet")
}
