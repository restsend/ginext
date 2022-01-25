package ginext

import (
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func addUser(t *testing.T, client *TestHTTPClient, r *gin.Engine, username, email, password string) {
	form := RegisterUserForm{
		UserName: username,
		Email:    email,
		Password: password,
	}
	var info UserInfoResult
	err := client.Call("/auth/register", &form, &info)
	assert.Nil(t, err)
	assert.Equal(t, username, info.UserName)
}

func TestUserRegister(t *testing.T) {
	um, r := NewTestUserManager()
	um.RegisterHandler("/auth", r)
	client := NewTestHTTPClient(r)
	addUser(t, client, r, "bob", "bob@example.org", "123456")
	{
		form := RegisterUserForm{
			UserName: "bob",
			Email:    "bob@example.org",
			Password: "123456",
		}

		var info UserInfoResult
		err := client.Call("/auth/register", &form, &info)
		assert.NotNil(t, err)
		assert.Equal(t, err.Error(), "username is exists")
	}
	{
		form := RegisterUserForm{
			UserName:    "alice",
			Email:       "alice@example.org",
			Password:    "123456",
			DisplayName: "AliceD",
			FirstName:   "AliceF",
			LastName:    "AliceL",
			Timezone:    "-480",
			Locale:      "zh-CN",
			Source:      "unittest",
		}
		var r UserInfoResult
		err := client.Call("/auth/register", &form, &r)
		assert.Nil(t, err)

		loginForm := LoginForm{
			UserName: "alice",
			Password: "123456",
		}
		var loginR UserInfoResult
		err = client.Call("/auth/login", &loginForm, &loginR)
		assert.Nil(t, err)

		var profile UserProfileResult
		err = client.Call("/auth/profile", nil, &profile)
		assert.Nil(t, err)
		assert.Equal(t, profile.DisplayName, form.DisplayName)
		assert.Equal(t, profile.FirstName, form.FirstName)
		assert.Equal(t, profile.LastName, form.LastName)
		assert.Equal(t, profile.Timezone, form.Timezone)
		assert.Equal(t, profile.Locale, form.Locale)
		u, err := um.GetByEmail("alice@example.org")
		assert.Nil(t, err)
		assert.Equal(t, u.Source, form.Source)
	}
}

func TestUserLogin(t *testing.T) {
	um, r := NewTestUserManager()
	um.RegisterHandler("/auth", r)

	client := NewTestHTTPClient(r)
	addUser(t, client, r, "bob", "bob@example.org", "123456")

	{
		form := LoginForm{
			UserName: "bob",
			Password: "123456",
		}
		var loginR UserInfoResult
		err := client.Call("/auth/login", &form, &loginR)
		assert.Nil(t, err)
		assert.Equal(t, loginR.UserName, form.UserName)
	}
	{
		form := RegisterUserForm{
			UserName: "bob",
			Password: "123",
		}
		var info UserInfoResult
		err := client.Call("/auth/login", &form, &info)
		assert.NotNil(t, err)
		assert.Equal(t, err.Error(), "bad password")
	}
}

func TestLoginSession(t *testing.T) {
	um, r := NewTestUserManager()
	um.RegisterHandler("/auth", r)

	client := NewTestHTTPClient(r)
	addUser(t, client, r, "bob", "bob@example.org", "123456")

	{
		form := LoginForm{
			UserName: "bob",
			Password: "123456",
		}
		var loginR UserInfoResult
		err := client.Call("/auth/login", &form, &loginR)
		assert.Nil(t, err)

		r.POST("/current", func(c *gin.Context) {
			curretUser := CurrentUser(c)
			if curretUser == nil {
				c.JSON(200, gin.H{
					"username": "BAD SESSION",
				})
			} else {
				c.JSON(200, gin.H{
					"username": curretUser.UserName,
				})
			}
		})

		w := client.Post("/current", nil)
		resp := client.CheckResponse(t, w)
		assert.NotNil(t, resp)
		assert.Equal(t, resp["username"], "bob")

		{
			result := um.db.Model(&GinExtUser{}).Where("user_name", "bob").UpdateColumn("Enabled", false)
			assert.Equal(t, result.RowsAffected, int64(1))
			w = client.Post("/current", nil)
			resp = client.CheckResponse(t, w)
			assert.NotNil(t, resp)
			assert.Equal(t, resp["username"], "BAD SESSION")
			um.db.Model(&GinExtUser{}).Where("user_name", "bob").UpdateColumn("Enabled", true)
		}
		w = client.Post("/auth/logout", nil)
		resp = client.CheckResponse(t, w)
		assert.NotNil(t, resp)

		w = client.Post("/current", nil)
		resp = client.CheckResponse(t, w)
		assert.NotNil(t, resp)
		assert.Equal(t, resp["username"], "BAD SESSION")
	}
}

func TestLoginToken(t *testing.T) {
	um, r := NewTestUserManager()
	um.RegisterHandler("/auth", r)

	client := NewTestHTTPClient(r)
	addUser(t, client, r, "bob", "bob@example.org", "123456")
	var token string
	{
		data := map[string]interface{}{
			"username": "bob",
			"password": "123456",
		}
		w := client.Post("/auth/token", data)
		resp := client.CheckResponse(t, w)
		assert.NotNil(t, resp)
		resultData := resp["data"].(map[string]interface{})
		token = resultData["token"].(string)

		r.Any("/current", func(c *gin.Context) {
			curretUser := CurrentUser(c)
			if curretUser == nil {
				c.JSON(200, gin.H{
					"username": "BAD SESSION",
				})
			} else {
				c.JSON(200, gin.H{
					"username": curretUser.UserName,
				})
			}
		})

		req, _ := http.NewRequest("GET", "/current", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w = NewTestHTTPClient(r).SendReq("/current", req)
		resp = client.CheckResponse(t, w)
		assert.NotNil(t, resp)
		assert.Equal(t, "bob", resp["username"])

		w = NewTestHTTPClient(r).Post("/current", nil)
		resp = client.CheckResponse(t, w)
		assert.NotNil(t, resp)
		assert.Equal(t, "BAD SESSION", resp["username"])
	}
	{
		form := TokenRefreshForm{
			Token: token,
		}
		var result TokenResult
		err := client.Call("/auth/refresh", form, &result)
		assert.Nil(t, err)
		assert.Equal(t, result.Token, form.Token)
		assert.Less(t, time.Now().Unix(), result.ExpiredAt.Unix())
	}
}

func TestActivedLogin(t *testing.T) {
	um, r := NewTestUserManager()
	um.RegisterHandler("/auth", r)

	client := NewTestHTTPClient(r)
	addUser(t, client, r, "bob", "bob@example.org", "123456")
	um.ext.SetValue(key_ACTIVE_REQUIRED, "true")
	{
		loginForm := LoginForm{
			UserName: "bob",
			Password: "123456",
		}
		var loginResult UserInfoResult
		err := client.Call("/auth/login", loginForm, &loginResult)
		assert.NotNil(t, err)
		assert.EqualError(t, err, "user need actived first")
	}
	{
		loginForm := LoginForm{
			UserName: "bob",
			Password: "123456",
		}
		var tokenResult TokenResult
		err := client.Call("/auth/token", loginForm, &tokenResult)
		assert.NotNil(t, err)
		assert.EqualError(t, err, "user need actived first")
	}
}
func TestVerifyRegisterEmail(t *testing.T) {
	um, r := NewTestUserManager()
	um.RegisterHandler("/auth", r)

	client := NewTestHTTPClient(r)
	addUser(t, client, r, "bob", "bademail@unittest", "123456")
	{
		form := VerifyEmailForm{
			Email: "alice@example.org",
		}
		key := ""
		err := client.Call("/auth/verifyemail/new", &form, &key)
		assert.Nil(t, err)
	}
	{
		form := RegisterUserForm{
			Email:    "alice@example.org",
			Password: "password",
		}
		var info UserInfoResult
		err := client.Call("/auth/register", &form, &info)
		assert.NotNil(t, err)
	}
}

func TestVerifyBindEmail(t *testing.T) {
	um, r := NewTestUserManager()
	um.RegisterHandler("/auth", r)

	client := NewTestHTTPClient(r)
	addUser(t, client, r, "bob", "bademail@unittest", "123456")
	code := ""
	Sig().Connect(SigUserVerifyEmail, func(sender interface{}, params ...interface{}) {
		code = params[1].(string)
	})

	{
		form := VerifyEmailForm{
			Email: "bob@example.org",
		}
		key := ""
		err := client.Call("/auth/verifyemail", &form, &key)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "Unauthorized")
	}
	{
		form := LoginForm{
			UserName: "bob",
			Password: "123456",
		}
		var loginR UserInfoResult
		err := client.Call("/auth/login", &form, &loginR)
		assert.Nil(t, err)
	}
	{
		form := VerifyEmailForm{
			Email: "bob@example.org",
		}
		key := ""
		err := client.Call("/auth/verifyemail", &form, &key)
		assert.Nil(t, err)
		assert.NotEmpty(t, key)

		bform := BindEmailForm{
			Email:    "bob@example.org",
			Password: "778899",
			Key:      key,
			Code:     code,
		}
		r := false
		err = client.Call("/auth/bindemail", &bform, &r)
		assert.Nil(t, err)
		assert.True(t, r)
	}
	{
		form := LoginForm{
			UserName: "bob",
			Password: "123456",
		}
		var loginR UserInfoResult
		err := client.Call("/auth/login", &form, &loginR)
		assert.NotNil(t, err)

		form.Password = "778899"
		err = client.Call("/auth/login", &form, &loginR)
		assert.Nil(t, err)
	}
}

func TestResetpassword(t *testing.T) {
	um, r := NewTestUserManager()
	um.RegisterHandler("/auth", r)

	client := NewTestHTTPClient(r)
	addUser(t, client, r, "bob", "bob@example.org", "123456")
	code := ""
	Sig().Connect(SigUserResetpassword, func(sender interface{}, params ...interface{}) {
		code = params[1].(string)
	})
	{
		form := PasswordLostForm{
			Email: "bob@example.org",
		}
		key := ""
		err := client.Call("/auth/password/lost", &form, &key)
		assert.Nil(t, err)
		assert.NotEmpty(t, key)

		bform := PasswordResetForm{
			Email:    "bob@example.org",
			Password: "778899",
			Key:      key,
			Code:     code,
		}
		r := false
		err = client.Call("/auth/password/reset", &bform, &r)
		assert.Nil(t, err)
		assert.True(t, r)
	}
	{
		form := LoginForm{
			UserName: "bob",
			Password: "123456",
		}
		var loginR UserInfoResult
		err := client.Call("/auth/login", &form, &loginR)
		assert.NotNil(t, err)

		form.Password = "778899"
		err = client.Call("/auth/login", &form, &loginR)
		assert.Nil(t, err)
	}
}

func TestChangepassword(t *testing.T) {
	um, r := NewTestUserManager()
	um.RegisterHandler("/auth", r)
	client := NewTestHTTPClient(r)
	addUser(t, client, r, "bob", "bob@example.org", "123456")

	{
		form := LoginForm{
			UserName: "bob",
			Password: "123456",
		}
		var loginR UserInfoResult
		err := client.Call("/auth/login", &form, &loginR)
		assert.Nil(t, err)
	}
	{
		form := PasswordChangeForm{
			Password: "778899",
		}
		r := false
		err := client.Call("/auth/password/change", &form, &r)
		assert.Nil(t, err)
		assert.True(t, r)
	}
	{
		form := LoginForm{
			UserName: "bob",
			Password: "123456",
		}
		var loginR UserInfoResult
		err := client.Call("/auth/login", &form, &loginR)
		assert.NotNil(t, err)

		form.Password = "778899"
		err = client.Call("/auth/login", &form, &loginR)
		assert.Nil(t, err)
	}
}
