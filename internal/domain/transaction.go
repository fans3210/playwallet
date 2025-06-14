package domain

import (
	"fmt"
	"strings"
	"time"

	"playwallet/pkg/errs"
)

type TransactionType string

const (
	Deposit  TransactionType = "deposit"
	Withdraw TransactionType = "withdraw"
	Transfer TransactionType = "transfer"
)

func (t TransactionType) IsValid() bool {
	return t == Deposit || t == Withdraw || t == Transfer
}

type TransactionReq struct {
	IdempotencyKey string          `json:"idempotency_key"`
	UserID         int64           `json:"userid"`
	TargetID       *int64          `json:"targetid"` // only passed when type is `transfer`
	Amt            int64           `json:"amt"`
	Type           TransactionType `json:"type"`
}

func (t TransactionReq) Validate() error {
	errMsgs := make([]string, 0, 5)
	if strings.TrimSpace(t.IdempotencyKey) == "" {
		errMsgs = append(errMsgs, "empty idempotency key")
	}
	if t.UserID <= 0 {
		errMsgs = append(errMsgs, "invalid userid")
	}
	if t.Amt <= 0 {
		errMsgs = append(errMsgs, "amount val <= 0")
	}
	if !t.Type.IsValid() {
		errMsgs = append(errMsgs, fmt.Sprintf("transaction type: %s not supported", t.Type))
	}
	if t.Type == Transfer {
		if t.TargetID == nil {
			errMsgs = append(errMsgs, "target id not specified")
		} else if *t.TargetID <= 0 {
			errMsgs = append(errMsgs, "invalid target id ")
		}
	}
	if len(errMsgs) > 0 {
		return errs.ValidationErrWithReason(errMsgs...)
	}
	return nil
}

// primaryKey = ID(IdempotencyKey) + UserID, because each transfer would create two transactions, both have same IdempotencyKey but different userid
type Transaction struct {
	IdempotencyKey string    `gorm:"column:idempotencykey;primaryKey;check:idempotencykey<>''"` // WARN: transction id is not unique, transctionid+userid is unique, refers to IdempotencyKey of FrozenBalance for `transfer` case,
	UserID         int64     `gorm:"column:userid;primaryKey;check:userid>0;index"`
	TargetID       *int64    `gorm:"column:targetid;check:targetid is null or targetid > 0"` // if not speicying targetid, the transaction would be credit or debit , otherwise, is a transfer
	Amount         int64     `gorm:"column:amt;not null;check:amt>0"`
	IsDebit        bool      `gorm:"column:isdebit;not null"`
	At             time.Time `gorm:"column:at;not null"`
}

// each fronzen balance map to 2 transaction records => one debit, one credit
func FrozenBalancesToTransactions(at time.Time, fbs ...FrozenBalance) []Transaction {
	ret := make([]Transaction, 0, 2*len(fbs))
	for _, fb := range fbs {
		if fb.Status == FrozenStatusConfirmed {
			continue
		}
		transDebit := Transaction{
			IdempotencyKey: fb.IdempotencyKey,
			UserID:         fb.UserID,
			TargetID:       &fb.TargetID,
			Amount:         fb.Amount,
			IsDebit:        true,
			At:             at,
		}
		transCredit := Transaction{
			IdempotencyKey: fb.IdempotencyKey,
			UserID:         fb.TargetID,
			TargetID:       &fb.UserID,
			Amount:         fb.Amount,
			IsDebit:        false,
			At:             at,
		}
		ret = append(ret, transDebit, transCredit)
	}
	return ret
}
