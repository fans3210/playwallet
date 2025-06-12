package biz

import (
	"fmt"

	"playwallet/internal/domain"
	"playwallet/pkg/errs"
)

func (uc *WalletUC) Deposit(req domain.DepositReq) error {
	if !uc.repo.CheckUserExist(req.UserID) {
		return fmt.Errorf("user not exist, %w", errs.ErrNotFound)
	}
	return uc.repo.Deposit(req)
}
