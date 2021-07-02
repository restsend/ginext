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
	errServerError
)

const defaultTokenExpired = 7 * 86400 * time.Second
const defaultTokenLength = 24

type UserManager struct {
	db           *gorm.DB
	PasswordSalt string
	TokenExpired time.Duration
	TokenLength  int
}

func NewUserManager(db *gorm.DB) *UserManager {
	return &UserManager{
		db:           db,
		PasswordSalt: "",
		TokenExpired: defaultTokenExpired,
		TokenLength:  defaultTokenLength,
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
	um.db.Model(user).Updates(GinExtUser{LastLoginIP: lastIp, LastLogin: sql.NullTime{Time: time.Now(), Valid: true}})
}
