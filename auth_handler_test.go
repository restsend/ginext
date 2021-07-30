package ginext

import (
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func addUser(t *testing.T, client *TestHTTPClient, r *gin.Engine, username, email, password string) {
	data := map[string]interface{}{
		"username": username,
		"email":    email,
		"password": password,
	}

	w := client.Post("/auth/register", data)
	resp := client.CheckResponse(t, w)
	assert.NotNil(t, resp)
	resultData := resp["data"].(map[string]interface{})
	assert.Equal(t, resultData["username"], "bob")
}

func TestUserRegister(t *testing.T) {
	um, r := NewTestUserManager()
	um.RegisterHandler("/auth", r)
	client := NewTestHTTPClient(r)
	addUser(t, client, r, "bob", "bob@example.org", "123456")
	{
		data := map[string]interface{}{
			"username": "bob",
			"email":    "bob@example.org",
			"password": "123456",
		}

		client := NewTestHTTPClient(r)
		w := client.Post("/auth/register", data)
		resp := client.CheckResponse(t, w)
		assert.NotNil(t, resp)
		assert.Equal(t, resp["msg"], "username is exists")
	}
}

func TestUserLogin(t *testing.T) {
	um, r := NewTestUserManager()
	um.RegisterHandler("/auth", r)

	client := NewTestHTTPClient(r)
	addUser(t, client, r, "bob", "bob@example.org", "123456")

	{
		data := map[string]interface{}{
			"username": "bob",
			"password": "123456",
		}
		w := client.Post("/auth/login", data)
		resp := client.CheckResponse(t, w)
		assert.NotNil(t, resp)
		resultData := resp["data"].(map[string]interface{})
		assert.Equal(t, resultData["username"], "bob")
	}
	{
		data := map[string]interface{}{
			"username": "bob",
			"password": "123",
		}
		w := client.Post("/auth/login", data)
		resp := client.CheckResponse(t, w)
		assert.NotNil(t, resp)
		assert.Equal(t, resp["msg"], "bad password")
	}
}

func TestLoginSession(t *testing.T) {
	um, r := NewTestUserManager()
	um.RegisterHandler("/auth", r)

	client := NewTestHTTPClient(r)
	addUser(t, client, r, "bob", "bob@example.org", "123456")

	{
		data := map[string]interface{}{
			"username": "bob",
			"password": "123456",
		}
		w := client.Post("/auth/login", data)
		resp := client.CheckResponse(t, w)
		assert.NotNil(t, resp)
		resultData := resp["data"].(map[string]interface{})
		assert.Equal(t, resultData["username"], "bob")

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

		w = client.Post("/current", data)
		resp = client.CheckResponse(t, w)
		assert.NotNil(t, resp)
		assert.Equal(t, resp["username"], "bob")

		w = client.Post("/auth/logout", data)
		resp = client.CheckResponse(t, w)
		assert.NotNil(t, resp)

		w = client.Post("/current", data)
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
