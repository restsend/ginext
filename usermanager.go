package ginext

import (
	"crypto/md5"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"gorm.io/gorm"
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
const defaultActiveRequired = false
const key_ACTIVE_REQUIRED = "GINEXT_ACTIVE_REQUIRED"

type UserManager struct {
	ext            *GinExt
	db             *gorm.DB
	PasswordSalt   string
	TokenExpired   time.Duration
	TokenLength    int
	ActiveRequired bool
}

func NewUserManager(ext *GinExt) *UserManager {
	return &UserManager{
		ext:            ext,
		db:             ext.DbInstance,
		PasswordSalt:   "",
		TokenExpired:   defaultTokenExpired,
		TokenLength:    defaultTokenLength,
		ActiveRequired: defaultActiveRequired,
	}
}

func (um *UserManager) Init() (err error) {
	err = um.db.AutoMigrate(&GinExtUser{})
	if err != nil {
		log.Panicf("Migrate GinExtUser Fail %v", err)
		return err
	}
	err = um.db.AutoMigrate(&GinToken{})
	if err != nil {
		log.Panicf("Migrate GinToken Fail %v", err)
		return err
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
	vals := map[string]interface{}{
		"LastLoginIP": lastIp,
		"LastLogin":   sql.NullTime{Time: time.Now(), Valid: true},
	}
	um.db.Model(user).Updates(vals)
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
