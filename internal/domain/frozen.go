package domain

import "time"

type FrozenStatus int

const (
	FrozenStatusFrozen FrozenStatus = iota + 1
	FrozenStatusConfirmed
)

// FrozenBalance record would only be created if there is a transfer event
type FrozenBalance struct {
	IdempotencyKey string       `gorm:"column:idempotencykey;primaryKey;check:idempotencykey <> ''"`
	UserID         int64        `gorm:"column:userid;not null;index;check:userid > 0"`
	TargetID       int64        `gorm:"column:targetid;not null;check:targetid > 0"`
	Amount         int64        `gorm:"column:amt;not null;check:amt>0"`
	Status         FrozenStatus `gorm:"column:status;check:status in (1,2);not null"`
	At             time.Time    `gorm:"column:at;not null"`
}
