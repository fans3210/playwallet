package domain

type User struct {
	ID       int64  `gorm:"primaryKey;check:id>0"`
	UserName string `gorm:"column:username;uniqueIndex;check:username<>''"`
}
