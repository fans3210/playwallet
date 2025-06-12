package biz

import (
	"fmt"

	"playwallet/internal/domain"
	"playwallet/pkg/errs"
)

func (uc *WalletUC) CheckBalance(userID int64) (*domain.BalanceInfo, error) {
	if !uc.repo.CheckUserExist(userID) {
		return nil, fmt.Errorf("user not exist, %w", errs.ErrNotFound)
	}
	baseInfo, err := uc.repo.CheckBalance(userID)
	if err != nil {
		return nil, err
	}
	availableBalance := max(baseInfo.TotalBalance-baseInfo.FrozenBalance, 0)
	return &domain.BalanceInfo{
		BalanceBaseInfo:  *baseInfo,
		UserID:           userID,
		AvailableBalance: availableBalance,
		Unit:             "cents",
	}, nil
}
