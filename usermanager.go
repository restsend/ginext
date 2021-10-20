package ginext

import (
	"crypto/md5"
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	errUsernameExists = 10000 + iota
	errEmailExists
	errBadPassword
	errInvalidParams
	errNotAllowed
	errActiveRequired
	errServerError
)

const defaultTokenExpired = 7 * 86400 * time.Second
const defaultTokenLength = 24
const key_ACTIVE_REQUIRED = "GINEXT_ACTIVE_REQUIRED"
const defaultCodeExpired = 180 * time.Second
const defaultMaxVerifyFailCount = 5
const defaultVerifyCodeLength = 6

type UserManager struct {
	ext          *GinExt
	db           *gorm.DB
	PasswordSalt string
	TokenExpired time.Duration
	TokenLength  int
}

func NewUserManager(ext *GinExt) *UserManager {
	return &UserManager{
		ext:          ext,
		db:           ext.DbInstance,
		PasswordSalt: "",
		TokenExpired: defaultTokenExpired,
		TokenLength:  defaultTokenLength,
	}
}

func (um *UserManager) Init() (err error) {
	tables := []interface{}{
		&GinExtUser{},
		&GinToken{},
		&GinProfile{},
		&GinVerifyCode{},
	}
	for _, t := range tables {
		err = um.db.AutoMigrate(t)
		if err != nil {
			log.Panicf("Migrate %s Fail %v", reflect.TypeOf(t).Name(), err)
			return err
		}
	}
	um.ext.CheckValue(key_ACTIVE_REQUIRED, "false")
	return nil
}

func (um *UserManager) Create(username, email, password string) (user *GinExtUser, err error) {
	if len(username) <= 0 {
		return nil, errors.New("empty username")
	}

	user = &GinExtUser{
		UserName: strings.ToLower(username),
		Email:    email,
		Password: um.hashPassword(password),
		Enabled:  true,
		Actived:  false,
	}

	result := um.db.Create(&user)
	return user, result.Error
}

func (um *UserManager) Get(username string) (user *GinExtUser, err error) {
	result := um.db.Where("user_name", strings.ToLower(username)).Take(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return user, nil
}

func (um *UserManager) GetByEmail(email string) (user *GinExtUser, err error) {
	result := um.db.Where("email", strings.ToLower(email)).Take(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return user, nil
}

func (um *UserManager) Auth(usernameOrEmail, rawPassword string) (user *GinExtUser, err error) {
	lowerVal := strings.ToLower(usernameOrEmail)
	//var userObj User
	result := um.db.Where("user_name", lowerVal).Or("email", lowerVal).Take(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	hashVal := um.hashPassword(rawPassword)
	if user.Password != hashVal {
		return nil, errors.New("bad password")
	}

	if !user.Enabled {
		return nil, errors.New("user is not allow login")
	}

	if !um.CheckForceActived(user) {
		return nil, errors.New("user need actived first")
	}

	return user, nil
}

func (um *UserManager) IsExists(username string) bool {
	_, err := um.Get(username)
	return err == nil
}

func (um *UserManager) IsExistsByEmail(email string) bool {
	_, err := um.GetByEmail(email)
	return err == nil
}

func (um *UserManager) SetPassword(user *GinExtUser, password string) {
	hashVal := um.hashPassword(password)
	um.db.Model(user).Update("password", hashVal)
}

func (um *UserManager) hashPassword(password string) string {
	hashVal := md5.Sum([]byte(um.PasswordSalt + password))
	return fmt.Sprintf("md5$%s%x", um.PasswordSalt, hashVal)
}

func (um *UserManager) SetLastLogin(user *GinExtUser, lastIp string) {
	now := time.Now()
	vals := map[string]interface{}{
		"LastLoginIP": lastIp,
		"LastLogin":   &now,
	}
	user.LastLogin = &now
	user.LastLoginIP = lastIp
	um.db.Model(user).Updates(vals)
}

func (um *UserManager) SetEmail(user *GinExtUser, val string) {
	um.db.Model(user).Updates(map[string]interface{}{"Email": val})
}

func (um *UserManager) SetActived(user *GinExtUser, val bool) {
	um.db.Model(user).Updates(map[string]interface{}{"Actived": val})
}

func (um *UserManager) SetEnabled(user *GinExtUser, val bool) {
	um.db.Model(user).Updates(map[string]interface{}{"Enabled": val})
}

func (um *UserManager) SetIsStaff(user *GinExtUser, val bool) {
	um.db.Model(user).Updates(map[string]interface{}{"IsStaff": val})
}

func (um *UserManager) SetPhone(user *GinExtUser, val string) {
	um.db.Model(user).Updates(map[string]interface{}{"Phone": val})
}

func (um *UserManager) SetName(user *GinExtUser, firstName, lastName string) {
	um.db.Model(user).Updates(map[string]interface{}{"FirstName": firstName, "LastName": lastName})
}

func (um *UserManager) SetDisplayName(user *GinExtUser, val string) {
	um.db.Model(user).Updates(map[string]interface{}{"DisplayName": val})
}

func (um *UserManager) SetSource(user *GinExtUser, val string) {
	if len(user.Source) <= 0 {
		um.db.Model(user).Updates(map[string]interface{}{"Source": val})
	}
}

func (um *UserManager) CheckForceActived(user *GinExtUser) bool {
	if user.Actived {
		return true
	}

	val := strings.ToLower(um.ext.GetValue(key_ACTIVE_REQUIRED))
	if val == "true" || val == "1" {
		if !user.Actived {
			return false
		}
	}
	return true
}

func (um *UserManager) genVerifyCode(user *GinExtUser, email string) (string, string) {
	key := RandText(20)
	if user != nil {
		key += fmt.Sprintf("-%d", user.ID)
	}
	code := RandNumberText(defaultVerifyCodeLength)
	val := GinVerifyCode{
		Key:       key,
		Source:    email,
		Code:      code,
		FailCount: 0,
		Verified:  false,
		ExpiredAt: time.Now().Add(defaultCodeExpired),
	}
	result := um.db.Create(&val)
	if result.Error != nil {
		return "", ""
	}
	return key, code
}

func (um *UserManager) verifyCode(key, email, code string) bool {
	var val GinVerifyCode
	result := um.db.Where("key", key).Take(&val)
	if result.Error != nil {
		return false
	}
	if time.Since(val.ExpiredAt) > 0 {
		return false
	}
	if val.FailCount > defaultMaxVerifyFailCount {
		return false
	}
	if val.Verified {
		return false
	}
	if val.Code != code || val.Source != email {
		um.db.Model(&val).UpdateColumn("fail_count", val.FailCount+1)
		return false
	}
	um.db.Delete(&val)
	return true
}

// Profile
func UpdateProfile(db *gorm.DB, userId uint, profile *GinProfile) (p *GinProfile, err error) {
	vals := map[string]interface{}{
		"Avatar":   profile.Avatar,
		"Gender":   profile.Gender,
		"Province": profile.Province,
		"City":     profile.City,
		"Country":  profile.Country,
		"Locale":   profile.Locale,
		"Timezone": profile.Timezone,
	}
	val := *profile
	result := db.Clauses(
		clause.OnConflict{
			Columns: []clause.Column{
				{Name: "user_id"},
			},
			DoUpdates: clause.Assignments(vals),
		}).Create(&val)
	if result.RowsAffected > 0 {
		db.Take(&val.User, val.UserID)
	}
	return &val, result.Error
}

func GetProfile(db *gorm.DB, userId uint) (profile *GinProfile, err error) {
	var val GinProfile
	result := db.Where("user_id", userId).Preload("User").FirstOrCreate(&val)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected > 0 {
		db.Take(&val.User, val.UserID)
	}
	return &val, nil
}
