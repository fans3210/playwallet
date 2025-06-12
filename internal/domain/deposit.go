package domain

type DepositReq struct {
	UserID         int64  `json:"userid"`
	Amt            int64  `json:"amt"`
	IdempotencyKey string `json:"idempotency_key"`
}
