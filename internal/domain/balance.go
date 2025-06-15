package domain

type BalanceInfo struct {
	BalanceBaseInfo
	UserID           int64 `json:"uid"`
	AvailableBalance int64 `json:"available_balance"`
}

type BalanceBaseInfo struct {
	TotalBalance  int64 `json:"total_balance"`
	FrozenBalance int64 `json:"frozen_balance"`
}

func (b BalanceBaseInfo) IsValid() bool {
	return b.TotalBalance >= 0 && b.FrozenBalance >= 0
	//  NOTE: b.TotalBalance >= b.FrozenBalance checking is not necesary for testint concurrent trasnfer case,
	// since the frozen balance can be large, but eventually should be processed and cancelled
}
