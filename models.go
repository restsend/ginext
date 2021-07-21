package ginext

import (
	"database/sql"
	"time"
)

type GinExtUser struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time

	UserName    string `gorm:"size:128;uniqueIndex"`
	Email       string `gorm:"size:128;uniqueIndex"`
	Phone       string `gorm:"size:64;index"`
	FirstName   string `gorm:"size:128"`
	LastName    string `gorm:"size:128"`
	Password    string `gorm:"size:128"`
	DisplayName string `gorm:"size:128"`
	IsStaff     bool
	Enabled     bool
	LastLogin   sql.NullTime
	LastLoginIP string `gorm:"size:128"`
}

type GinExtConfig struct {
	ID    uint   `gorm:"primarykey"`
	Key   string `gorm:"size:128;uniqueIndex"`
	Value string
}

type GinTask struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	TaskType  string `gorm:"size:64"`
	// For example: shop.id, user.id
	ObjectID int64 `gorm:"index"`
	Done     bool  `gorm:"index"`
	Failed   bool  `gorm:"index"`
	Context  string
	Result   string
	// Delay to invoke
	StartTime sql.NullTime `gorm:"index"`
	ExecTime  sql.NullTime
	EndTime   sql.NullTime
}

type GinToken struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	OwnerID   uint
	Owner     GinExtUser
	Token     string `gorm:"size:32;uniqueIndex"`
	ExpiredAt time.Time
}
