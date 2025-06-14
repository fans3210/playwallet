package biz

import (
	"fmt"

	"playwallet/internal/domain"
	"playwallet/pkg/errs"
)

func (uc *WalletUC) CheckBalance(userID int64) (*domain.BalanceInfo, error) {
	exists, err := uc.repo.CheckUserExist(userID)
	if err != nil {
		return nil, fmt.Errorf("check balance, check user %d exist err, %w", userID, err)
	}
	if !exists {
		return nil, fmt.Errorf("check balance, user %d not exist, %w", userID, errs.ErrNotFound)
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
	}, nil
}
