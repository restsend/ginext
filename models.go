package ginext

import (
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
	Actived     bool
	LastLogin   *time.Time
	LastLoginIP string `gorm:"size:128"`

	Source    string `gorm:"size:64;index"`
	WxOpenID  string `gorm:"size:100;"`
	WxUnionID string `gorm:"size:100;"`
	FBAuthID  string `gorm:"size:100;"`
	GGAuthID  string `gorm:"size:100;"`
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
	StartTime *time.Time `gorm:"index"`
	ExecTime  *time.Time
	EndTime   *time.Time
}

type GinToken struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	OwnerID   uint
	Owner     GinExtUser
	Token     string `gorm:"size:32;uniqueIndex"`
	ExpiredAt time.Time
}

type GinProfile struct {
	ID     uint       `json:"id" gorm:"primarykey"`
	UserID uint       `json:"userId" gorm:"uniqueIndex"`
	User   GinExtUser `json:"-"`

	Avatar   string `json:"avatar" gorm:"size:1024;"`
	Gender   string `json:"gender" gorm:"size:12;"`
	Province string `json:"province" gorm:"size:32;"`
	City     string `json:"city" gorm:"size:32;"`
	Country  string `json:"country" gorm:"size:32;"`
	Locale   string `json:"locale" gorm:"size:8;"`
	Timezone string `json:"timezone" gorm:"size:64;"`
}

type GinVerifyCode struct {
	ID        uint   `gorm:"primarykey"`
	Key       string `gorm:"size:64;uniqueIndex"`
	Source    string `gorm:"size:200"`
	Code      string `gorm:"size:12"`
	FailCount int
	Verified  bool
	ExpiredAt time.Time
}

func (u *GinExtUser) GetVisibleName() string {
	if len(u.DisplayName) > 0 {
		return u.DisplayName
	}
	if len(u.FirstName) > 0 {
		return u.FirstName
	}
	return u.LastName
}
