package domain

type BalanceInfo struct {
	BalanceBaseInfo
	UserID           int64  `json:"uid"`
	AvailableBalance int64  `json:"available_balance"`
	Unit             string `json:"unit"`
}

type BalanceBaseInfo struct {
	TotalBalance  int64 `json:"total_balance"`
	FrozenBalance int64 `json:"frozen_balance"`
}

func (b BalanceBaseInfo) IsValid() bool {
	return b.TotalBalance >= 0 && b.FrozenBalance >= 0 && b.TotalBalance >= b.FrozenBalance
}
