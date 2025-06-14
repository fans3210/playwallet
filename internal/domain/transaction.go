package domain

import (
	"fmt"
	"strings"
	"time"

	"playwallet/pkg/errs"

	"github.com/segmentio/kafka-go"
	"github.com/vmihailenco/msgpack/v5"
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

// can be used to send transfer req and store ledger records
// for sender who send this req, targetid => receiverid
// for receiver who received this req, targetid = otherid = senderid
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

func (t TransactionReq) ToKafkaMsg() (*kafka.Message, error) {
	b, err := msgpack.Marshal(t)
	if err != nil {
		return nil, err
	}
	return &kafka.Message{
		Key:   []byte(t.IdempotencyKey),
		Value: b,
	}, nil
}

// primaryKey = ID(IdempotencyKey) + UserID, because each transfer would create two transactions, both have same IdempotencyKey but different userid
// TODO: rename to ledger record
type Transaction struct {
	IdempotencyKey string    `gorm:"column:idempotencykey;primaryKey;check:idempotencykey<>''"` // WARN: transction id is not unique, transctionid+userid is unique, refers to IdempotencyKey of FrozenBalance for `transfer` case,
	UserID         int64     `gorm:"column:userid;primaryKey;check:userid>0;index"`
	OtherID        *int64    `gorm:"column:otherid;check:otherid is null or otherid > 0"` // if not speicying otherid, the transaction would be credit or debit , otherwise, is a transfer
	Amount         int64     `gorm:"column:amt;not null;check:amt>0"`
	IsDebit        bool      `gorm:"column:isdebit;not null"`
	CreateAt       time.Time `gorm:"column:at;not null"`
}
