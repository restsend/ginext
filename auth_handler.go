package ginext

import (
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

/*
	/auth/register
	/auth/login
	/auth/refresh
	/auth/logout
	/auth/password/lost
	/auth/password/reset
*/

type RegisterUserForm struct {
	Email       string `json:"email" binding:"required"`
	Password    string `json:"password" binding:"required"`
	UserName    string `json:"username"`
	DisplayName string `json:"displayName"`
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	Locale      string `json:"locale"`
	Timezone    string `json:"timezone"`
	Code        string `json:"code"`
	Key         string `json:"key"`
	Source      string `json:"source"`
}

type LoginForm struct {
	UserName string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password" binding:"required"`
}

type TokenRefreshForm struct {
	Token string `json:"token"`
}

type PasswordLostForm struct {
	Email  string `json:"email" binding:"required"`
	Locale string `json:"locale"`
}

type PasswordChangeForm struct {
	Password string `json:"password" binding:"required"`
}

type PasswordResetForm struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
	Key      string `json:"key" binding:"required"`
	Code     string `json:"code" binding:"required"`
}

type BindEmailForm PasswordResetForm
type VerifyEmailForm PasswordLostForm

type UserInfoResult struct {
	UserName  string     `json:"username"`
	Email     string     `json:"email"`
	LastLogin *time.Time `json:"lastLogin,omitempty"`
}

type UserProfileResult struct {
	UserName    string `json:"username"`
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`

	LastLogin   *time.Time `json:"lastLogin,omitempty"`
	LastLoginIP string     `json:"lastLoginIp"`

	Avatar   string `json:"avatar"`
	Gender   string `json:"gender"`
	Province string `json:"province"`
	City     string `json:"city"`
	Country  string `json:"country"`
	Locale   string `json:"locale"`
	Timezone string `json:"timezone"`
}

type TokenResult struct {
	Token     string    `json:"token"`
	ExpiredAt time.Time `json:"expiredAt"`
}

func (um *UserManager) loadUMWithGin() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(UserMangerField, um)
		authValue := c.GetHeader("Authorization")
		if um.EnabledTokenAuthorization && len(authValue) > 0 {
			vals := strings.Split(authValue, " ")
			if len(vals) <= 1 || vals[0] != "Bearer" {
				//RpcFail(c, http.StatusBadRequest, "invalid token format")
				return
			}
			obj, err := um.GetUserByToken(vals[1])
			if err == nil && obj != nil {
				session := sessions.Default(c)
				userId := session.Get(UserIdField)
				if userId != obj.OwnerID {
					um.TouchToken(obj.ID)
					Login(c, &obj.Owner)
					c.Set(TokenField, obj)
				}
			} else {
				RpcFail(c, http.StatusBadRequest, "invalid accesstoken")
				return
			}
		}

		c.Next()
	}
}

const docRegister = `Register user`
const docLogin = `User login`
const docProfile = `User profile`
const docToken = `User get accesstoken with username and password`
const docRefresh = `User refresh AccessToken`
const docLogout = `User logout`
const docPasswordLost = `Find the password`
const docPasswordChange = `Change the password`
const docPasswordResetDone = `Reset the password`
const docBindEmail = `Bind email with password`
const docVerifyEmail = `Send verify code via email`

func (um *UserManager) RegisterHandler(prefix string, r *gin.Engine) {
	// Require session
	//
	r.Use(um.loadUMWithGin())

	AddDocAppLabel("User Auth")

	RpcDefine(r, &RpcContext{
		Form:         RegisterUserForm{},
		Result:       UserInfoResult{},
		OnlyPost:     true,
		RelativePath: filepath.Join(prefix, "/register"),
		Handler:      um.handleRegister,
		Doc:          docRegister,
	})
	RpcDefine(r, &RpcContext{
		Form:         LoginForm{},
		Result:       UserInfoResult{},
		OnlyPost:     true,
		RelativePath: filepath.Join(prefix, "/login"),
		Handler:      um.handleLogin,
		Doc:          docLogin,
	})
	RpcDefine(r, &RpcContext{
		AuthRequired: true,
		Result:       UserProfileResult{},
		RelativePath: filepath.Join(prefix, "/profile"),
		Handler:      um.handleProfile,
		Doc:          docProfile,
	})
	RpcDefine(r, &RpcContext{
		Form:         LoginForm{},
		Result:       TokenResult{},
		OnlyPost:     true,
		RelativePath: filepath.Join(prefix, "/token"),
		Handler:      um.handleToken,
		Doc:          docToken,
	})
	RpcDefine(r, &RpcContext{
		Form:         TokenRefreshForm{},
		OnlyPost:     true,
		RelativePath: filepath.Join(prefix, "/refresh"),
		Handler:      um.handleRefresh,
		Doc:          docRefresh,
	})
	RpcDefine(r, &RpcContext{
		RelativePath: filepath.Join(prefix, "/logout"),
		Handler:      um.handleLogout,
		Doc:          docLogout,
	})

	RpcDefine(r, &RpcContext{
		OnlyPost:     true,
		AuthRequired: true,
		Form:         VerifyEmailForm{},
		Result:       "",
		RelativePath: filepath.Join(prefix, "/verifyemail"),
		Handler:      um.handleVerifyEmail,
		Doc:          docVerifyEmail,
	})

	RpcDefine(r, &RpcContext{
		OnlyPost:     true,
		Form:         VerifyEmailForm{},
		Result:       "",
		RelativePath: filepath.Join(prefix, "/verifyemail/new"),
		Handler:      um.handleVerifyEmailNewUser,
		Doc:          docVerifyEmail,
	})

	RpcDefine(r, &RpcContext{
		OnlyPost:     true,
		AuthRequired: true,
		Form:         BindEmailForm{},
		Result:       true,
		RelativePath: filepath.Join(prefix, "/bindemail"),
		Handler:      um.handleBindEmail,
		Doc:          docBindEmail,
	})

	RpcDefine(r, &RpcContext{
		OnlyPost:     true,
		AuthRequired: true,
		Form:         PasswordChangeForm{},
		Result:       true,
		RelativePath: filepath.Join(prefix, "/password/change"),
		Handler:      um.handlePasswordChange,
		Doc:          docPasswordChange,
	})
	RpcDefine(r, &RpcContext{
		OnlyPost:     true,
		Form:         PasswordLostForm{},
		Result:       "",
		RelativePath: filepath.Join(prefix, "/password/lost"),
		Handler:      um.handlePasswordLost,
		Doc:          docPasswordLost,
	})
	RpcDefine(r, &RpcContext{
		OnlyPost:     true,
		Form:         PasswordResetForm{},
		Result:       true,
		RelativePath: filepath.Join(prefix, "/password/reset"),
		Handler:      um.handlePasswordReset,
		Doc:          docPasswordResetDone,
	})
}

//handleRegister User Register
func (um *UserManager) handleRegister(c *gin.Context) {
	form := c.MustGet(RpcFormField).(*RegisterUserForm)
	if form.UserName == "" {
		form.UserName = form.Email
	}

	if um.IsExists(form.UserName) {
		RpcFail(c, errUsernameExists, "username is exists")
		return
	}

	if um.IsExistsByEmail(form.Email) {
		RpcFail(c, errEmailExists, "email is exists")
		return
	}

	vals := map[string]interface{}{}
	if um.hasVerifyCode(form.Email) {
		if !um.verifyCode(form.Key, form.Email, form.Code) {
			RpcFail(c, errBadVerifyCode, "bad verifycode")
			return
		}
		vals["Actived"] = true
	}

	user, err := um.Create(form.UserName, form.Email, form.Password)
	if err != nil {
		RpcFail(c, errServerError, "create user fail")
		return
	}
	if len(form.DisplayName) > 0 {
		vals["DisplayName"] = form.DisplayName
	}
	if len(form.FirstName) > 0 {
		vals["FirstName"] = form.FirstName
	}
	if len(form.LastName) > 0 {
		vals["LastName"] = form.LastName
	}
	if len(form.Source) > 0 {
		vals["Source"] = form.Source
	}

	if len(vals) > 0 {
		um.db.Model(user).Updates(vals)
	}

	if len(form.Timezone) > 0 || len(form.Locale) > 0 {
		profile, err := GetProfile(um.db, user.ID)
		if err == nil {
			profile.Locale = form.Locale
			profile.Timezone = form.Timezone
			UpdateProfile(um.db, user.ID, profile)
		}
	}

	um.SetLastLogin(user, c.ClientIP())
	Sig().Emit(SigUserCreate, user, c)

	RpcOk(c, UserInfoResult{
		UserName:  user.UserName,
		Email:     user.Email,
		LastLogin: user.LastLogin,
	})
}

//handleRegister Login
func (um *UserManager) handleLogin(c *gin.Context) {
	form := c.MustGet(RpcFormField).(*LoginForm)
	if len(form.Email) <= 0 && len(form.UserName) <= 0 {
		RpcFail(c, errInvalidParams, "bad username or password")
		return
	}
	key := form.UserName
	if len(key) <= 0 {
		key = form.Email
	}
	user, err := um.Auth(key, form.Password)
	if err != nil {
		RpcFail(c, errInvalidParams, err.Error())
		return
	}

	// Login ..
	//
	Login(c, user)
	RpcOk(c, UserInfoResult{
		UserName:  user.UserName,
		Email:     user.Email,
		LastLogin: user.LastLogin,
	})
}

func (um *UserManager) handleProfile(c *gin.Context) {
	user := CurrentUser(c)
	profile, err := GetProfile(um.db, user.ID)
	if err != nil {
		RpcError(c, err)
		return
	}
	r := UserProfileResult{
		UserName:    user.UserName,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		LastLogin:   user.LastLogin,
		LastLoginIP: user.LastLoginIP,
		Avatar:      profile.Avatar,
		Gender:      profile.Gender,
		Province:    profile.Province,
		City:        profile.City,
		Country:     profile.Country,
		Locale:      profile.Locale,
		Timezone:    profile.Timezone,
	}
	RpcOk(c, r)
}

func (um *UserManager) handleToken(c *gin.Context) {
	form := c.MustGet(RpcFormField).(*LoginForm)
	if len(form.Email) <= 0 && len(form.UserName) <= 0 {
		RpcFail(c, errInvalidParams, "bad username or password")
		return
	}
	key := form.UserName
	if len(key) <= 0 {
		key = form.Email
	}
	user, err := um.Auth(key, form.Password)
	if err != nil {
		RpcFail(c, errInvalidParams, err.Error())
		return
	}

	token, err := um.MakeToken(user)
	if err != nil {
		RpcFail(c, errInvalidParams, "token build fail")
		return
	}

	// Login ..
	//
	Login(c, user)
	RpcOk(c, TokenResult{
		Token:     token.Token,
		ExpiredAt: token.ExpiredAt,
	})
}
func (um *UserManager) handleLogout(c *gin.Context) {
	Logout(c)
	RpcOk(c, gin.H{
		"result": true,
	})
}

func (um *UserManager) handleRefresh(c *gin.Context) {
	form := c.MustGet(RpcFormField).(*TokenRefreshForm)
	userToken, err := um.GetUserByToken(form.Token)
	if err != nil {
		RpcFail(c, errInvalidParams, err.Error())
		return
	}

	expire, err := um.TouchToken(userToken.ID)
	if err != nil {
		RpcFail(c, errInvalidParams, err.Error())
		return
	}

	RpcOk(c, TokenResult{
		Token:     userToken.Token,
		ExpiredAt: expire,
	})
}

func (um *UserManager) handleVerifyEmailNewUser(c *gin.Context) {
	form := c.MustGet(RpcFormField).(*VerifyEmailForm)
	_, err := um.GetByEmail(form.Email)
	if err == nil {
		RpcFail(c, errEmailExists, "Email has already exists")
		return
	}

	key, code := um.genVerifyCode(nil, form.Email)

	Sig().Emit(SigUserVerifyEmail, nil, form.Email, code, form.Locale)
	RpcOk(c, key)
}

func (um *UserManager) handleVerifyEmail(c *gin.Context) {
	user := CurrentUser(c)
	form := c.MustGet(RpcFormField).(*VerifyEmailForm)

	_, err := um.GetByEmail(form.Email)
	if err == nil {
		RpcOk(c, "")
		return
	}

	key, code := um.genVerifyCode(user, form.Email)

	Sig().Emit(SigUserVerifyEmail, user, form.Email, code, form.Locale)
	RpcOk(c, key)
}

func (um *UserManager) handleBindEmail(c *gin.Context) {
	user := CurrentUser(c)
	form := c.MustGet(RpcFormField).(*BindEmailForm)
	if !um.verifyCode(form.Key, form.Email, form.Code) {
		RpcOk(c, false)
		return
	}
	um.SetPassword(user, form.Password)
	um.SetEmail(user, form.Email)
	um.SetActived(user, true)
	RpcOk(c, true)
}

func (um *UserManager) handlePasswordChange(c *gin.Context) {
	user := CurrentUser(c)
	form := c.MustGet(RpcFormField).(*PasswordChangeForm)
	um.SetPassword(user, form.Password)
	RpcOk(c, true)
}

func (um *UserManager) handlePasswordLost(c *gin.Context) {
	form := c.MustGet(RpcFormField).(*PasswordLostForm)
	user, err := um.GetByEmail(form.Email)
	if err != nil {
		RpcOk(c, "")
		return
	}

	key, code := um.genVerifyCode(user, form.Email)
	Sig().Emit(SigUserResetpassword, user, form.Email, code, form.Locale)
	RpcOk(c, key)
}

func (um *UserManager) handlePasswordReset(c *gin.Context) {
	form := c.MustGet(RpcFormField).(*PasswordResetForm)
	user, err := um.GetByEmail(form.Email)
	if err != nil {
		RpcOk(c, false)
		return
	}
	if !um.verifyCode(form.Key, form.Email, form.Code) {
		RpcOk(c, false)
		return
	}
	um.SetPassword(user, form.Password)
	RpcOk(c, true)
}
