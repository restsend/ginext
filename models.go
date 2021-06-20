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
	FirstName   string `gorm:"size:128"`
	LastName    string `gorm:"size:128"`
	Password    string `gorm:"size:128"`
	IsStaff     bool
	Enabled     bool
	LastLogin   sql.NullTime
	LastLoginIP string
}

type GinExtConfig struct {
	ID    uint   `gorm:"primarykey"`
	Key   string `gorm:"size:128;uniqueIndex"`
	Value string `gorm:"size:1024"`
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
