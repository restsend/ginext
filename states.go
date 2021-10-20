package ginext

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

const (
	//SigUserLogin: user *GinExtUser, c *gin.Context
	SigUserLogin = "user.login"
	//SigUserLogout: user *GinExtUser, c *gin.Context
	SigUserLogout = "user.logout"
	//SigUserCreate: user *GinExtUser, c *gin.Context
	SigUserCreate = "user.create"
	//SigUserVerifyEmail: user *GinExtUser, email string , code string
	SigUserVerifyEmail = "user.verifyemail"
	//SigUserResetpassword: user *GinExtUser, email string , code string
	SigUserResetpassword = "user.resetpassword"
)

func Login(c *gin.Context, user *GinExtUser) {
	um := c.MustGet(UserMangerField).(*UserManager)
	um.SetLastLogin(user, c.ClientIP())
	session := sessions.Default(c)
	session.Set(UserIdField, user.ID)
	session.Save()
	Sig().Emit(SigUserLogin, user, c)
}

func CurrentUser(c *gin.Context) (user *GinExtUser) {
	if cachedObj, exists := c.Get(UserIdField); exists && cachedObj != nil {
		return cachedObj.(*GinExtUser)
	}

	session := sessions.Default(c)
	userId := session.Get(UserIdField)
	if userId == nil {
		return nil
	}

	um := c.MustGet(UserMangerField).(*UserManager)
	result := um.db.First(&user, userId)
	if result.Error != nil {
		return nil
	}
	c.Set(UserIdField, user)
	return user
}

func Logout(c *gin.Context) {
	c.Set(UserIdField, nil)
	session := sessions.Default(c)
	session.Delete(UserIdField)
	session.Save()
	Sig().Emit(SigUserLogin, nil, c)
}
