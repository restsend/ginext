package ginext

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func NewTestUserManager() (um *UserManager, r *gin.Engine) {
	cfg := NewGinExt("..")
	cfg.Init()

	um = NewUserManager(cfg.DbInstance)
	um.Init()
	um.db.Delete(&GinExtUser{}, "id > 0")
	r = gin.Default()
	cfg.WithGinExt(r)
	return um, r
}

func TestUserCreate(t *testing.T) {
	um, _ := NewTestUserManager()
	bob, err := um.Create("bob", "bob@example.org", "123456")
	assert.NoError(t, err)
	assert.Equal(t, bob.UserName, "bob")
}

func TestUserAuth(t *testing.T) {
	um, _ := NewTestUserManager()
	bob, err := um.Create("bob", "bob@example.org", "123456")
	assert.NoError(t, err)
	assert.Equal(t, bob.UserName, "bob")

	{
		bobAuth, err := um.Auth("bob", "123456")
		assert.NoError(t, err)
		assert.NotNil(t, bobAuth)
		assert.Equal(t, bobAuth.Email, "bob@example.org")
	}

	assert.False(t, um.IsExists("alice"))
	assert.False(t, um.IsExistsByEmail("alice@example.org"))
}

func TestHashPassword(t *testing.T) {
	um, _ := NewTestUserManager()
	hash1 := um.hashPassword("123456")
	hash2 := um.hashPassword("654321")
	assert.NotEqual(t, hash1, hash2)
}
func TestUserUpdatePassword(t *testing.T) {
	um, _ := NewTestUserManager()
	bob, _ := um.Create("bob", "bob@example.org", "123456")

	{
		bobAuth, _ := um.Auth("bob", "12")
		assert.Nil(t, bobAuth)
	}
	{
		um.SetPassword(bob, "654321")
		bobAuth, err := um.Auth("bob", "654321")
		assert.NoError(t, err)
		assert.NotNil(t, bobAuth)
	}
}
