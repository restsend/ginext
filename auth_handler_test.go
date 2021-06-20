package ginext

import (
	"testing"

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
		assert.Equal(t, resp["msg"], "Username is exists")
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
		assert.Equal(t, resp["msg"], "Bad password")
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
