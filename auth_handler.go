package ginext

import (
	"path/filepath"
	"time"

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
	UserName  string `json:"username" binding:"required"`
	Email     string `json:"email" binding:"required"`
	Password  string `json:"password" binding:"required"`
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
}

type LoginForm struct {
	UserName string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password" binding:"required"`
}

type PasswordResetForm struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password"`
	Code     string `json:"code"`
}

type UserInfoResult struct {
	UserName  string    `json:"username"`
	Email     string    `json:"email"`
	LastLogin time.Time `json:"lastLogin"`
}

func (um *UserManager) loadUMWithGin() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(UserMangerField, um)
		c.Next()
	}
}

const docRegister = `Register user`
const docLogin = `User login`
const docRefresh = `User refresh AccessToken`
const docLogout = `User logout`
const docPasswordLost = `Find the password`
const docPasswordResetDone = `Reset the password`

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
		Form:         PasswordResetForm{},
		RelativePath: filepath.Join(prefix, "/password/lost"),
		Handler:      um.handlePasswordLost,
		Doc:          docPasswordLost,
	})
	RpcDefine(r, &RpcContext{
		OnlyPost:     true,
		Form:         PasswordResetForm{},
		RelativePath: filepath.Join(prefix, "/password/reset"),
		Handler:      um.handlePasswordReset,
		Doc:          docPasswordResetDone,
	})
}

//handleRegister User Register
func (um *UserManager) handleRegister(c *gin.Context) {
	form := c.MustGet(RpcFormField).(*RegisterUserForm)
	if um.IsExists(form.UserName) {
		RpcFail(c, errUsernameExists, "Username is exists")
		return
	}

	if um.IsExistsByEmail(form.Email) {
		RpcFail(c, errEmailExists, "Email is exists")
		return
	}

	user, err := um.Create(form.UserName, form.Email, form.Password)
	if err != nil {
		RpcFail(c, errServerError, "Create user fail")
		return
	}

	um.SetLastLogin(user, c.ClientIP())

	RpcOk(c, UserInfoResult{
		UserName:  user.UserName,
		Email:     user.Email,
		LastLogin: user.LastLogin.Time,
	})
}

//handleRegister Login
func (um *UserManager) handleLogin(c *gin.Context) {
	form := c.MustGet(RpcFormField).(*LoginForm)
	if len(form.Email) <= 0 && len(form.UserName) <= 0 {
		RpcFail(c, errInvalidParams, "Bad username or password")
		return
	}
	key := form.UserName
	if len(key) <= 0 {
		key = form.Email
	}
	user, err := um.Auth(key, form.Password)
	if err != nil {
		RpcFail(c, errInvalidParams, "Bad password")
		return
	}

	// Login ..
	//
	Login(c, user)
	RpcOk(c, UserInfoResult{
		UserName:  user.UserName,
		Email:     user.Email,
		LastLogin: user.LastLogin.Time,
	})
}

func (um *UserManager) handleLogout(c *gin.Context) {
	Logout(c)
	RpcOk(c, gin.H{
		"result": true,
	})
}

func (um *UserManager) handleRefresh(c *gin.Context) {
}

func (um *UserManager) handlePasswordLost(c *gin.Context) {
}

func (um *UserManager) handlePasswordReset(c *gin.Context) {
}
