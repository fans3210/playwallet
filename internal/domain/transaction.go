package domain

import "time"

// primaryKey = ID(IdempotencyKey) + UserID, because each transfer would create two transactions, both have same IdempotencyKey but different userid
type Transaction struct {
	IdempotencyKey string    `gorm:"column:idempotencykey;primaryKey;check:idempotencykey<>''"` // WARN: transction id is not unique, transctionid+userid is unique, refers to IdempotencyKey of FrozenBalance for `transfer` case,
	UserID         int64     `gorm:"column:userid;primaryKey;check:userid>0"`
	OtherID        *int64    `gorm:"column:otherid;check:otherid is null or otherid>0"` // if not speicying OtherID, the transaction would be credit or debit , otherwise, is a transfer
	Amount         int64     `gorm:"column:amt;not null;check:amt>0"`                   // in cents
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
			OtherID:        &fb.OtherID,
			Amount:         fb.Amount,
			IsDebit:        true,
			At:             at,
		}
		transCredit := Transaction{
			IdempotencyKey: fb.IdempotencyKey,
			UserID:         fb.OtherID, // the receiver is the `OtherID` from a fronzen balance
			OtherID:        &fb.UserID,
			Amount:         fb.Amount,
			IsDebit:        false,
			At:             at,
		}
		ret = append(ret, transDebit, transCredit)
	}
	return ret
}
