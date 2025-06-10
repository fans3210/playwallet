package domain

type User struct {
	ID       int64  `gorm:"primaryKey;autoIncrement"`
	UserName string `gorm:"column:username;uniqueIndex"`
}

type FrozenStatus int

const (
	FrozenStatusFrozen FrozenStatus = iota + 1
	FrozenStatusConfirmed
)

type FrozenBalance struct {
	ID             int64        `gorm:"primaryKey;autoIncrement"`
	UserID         int64        `gorm:"column:userid;not null;index"`
	Amount         int64        `gorm:"column:amt;not null"` // in cents
	Status         FrozenStatus `gorm:"column:status;check:status in (1,2);not null"`
	IdempotencyKey string       `gorm:"column:idempotencykey;uniqueIndex;not null"`
}

// each transfer would create two transactions
type Transaction struct {
	ID      string `gorm:"primaryKey"` // refers to IdempotencyKey of FrozenBalance
	UserID  int64  `gorm:"column:userid;primaryKey"`
	OtherID int64  `gorm:"column:otherid"`
	Amount  int64  `gorm:"column:amt;not null"` // in cents
	IsDebit bool   `gorm:"column:isdebit;not null"`
}
