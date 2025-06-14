package domain

import "time"

type FrozenStatus string

const (
	FrozenStatusFrozen    FrozenStatus = "frozen"
	FrozenStatusConfirmed FrozenStatus = "confirmed"
	FrozenStatusCancelled FrozenStatus = "cancelled"
)

// FrozenBalance record would only be created if there is a transfer event
type FrozenBalance struct {
	IdempotencyKey string       `gorm:"column:idempotencykey;primaryKey;check:idempotencykey <> ''"`
	UserID         int64        `gorm:"column:userid;not null;index;check:userid > 0"`
	TargetID       int64        `gorm:"column:targetid;not null;check:targetid > 0"`
	Amount         int64        `gorm:"column:amt;not null;check:amt>0"`
	Status         FrozenStatus `gorm:"column:status;check:status in ('frozen','confirmed','cancelled');not null"`
	CreateAt       time.Time    `gorm:"column:at;not null"`
	UpdateAt       time.Time    `gorm:"column:update_at"`
}
