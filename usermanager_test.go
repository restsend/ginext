package ginext

import (
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func NewTestUserManager() (um *UserManager, r *gin.Engine) {
	cfg := NewGinExt("..")
	cfg.Init()

	um = NewUserManager(cfg)
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

func TestToken(t *testing.T) {
	um, _ := NewTestUserManager()
	bob, _ := um.Create("bob", "bob@example.org", "123456")

	{
		token, err := um.MakeToken(bob)
		assert.Nil(t, err)
		assert.LessOrEqual(t, defaultTokenLength, len(token.Token))
		u, err := um.GetUserByToken(token.Token)
		assert.Nil(t, err)
		assert.NotNil(t, u)
		assert.Equal(t, bob.UserName, u.Owner.UserName)

		_, err = um.GetUserByToken("bad")
		assert.NotNil(t, err)

		err = um.DeleteToken(token.Token)
		assert.Nil(t, err)

		_, err = um.GetUserByToken(token.Token)
		assert.NotNil(t, err)
	}
}
func TestBasicUpdate(t *testing.T) {
	um, _ := NewTestUserManager()
	bob, _ := um.Create("bob", "bob@example.org", "123456")

	um.SetIsStaff(bob, true)
	um.SetEnabled(bob, false)
	um.SetDisplayName(bob, "new name")
	um.SetName(bob, "xyz", "1234")
	um.SetPhone(bob, "+11234567")
	um.SetLastLogin(bob, "127.0.0.1")

	bob, _ = um.Get("bob")

	assert.Equal(t, true, bob.IsStaff)
	assert.Equal(t, false, bob.Enabled)
	assert.Equal(t, "new name", bob.DisplayName)
	assert.Equal(t, "xyz", bob.FirstName)
	assert.Equal(t, "1234", bob.LastName)
	assert.Equal(t, "127.0.0.1", bob.LastLoginIP)
	assert.LessOrEqual(t, 0*time.Second, time.Since(bob.LastLogin.Time))
}
func TestProfile(t *testing.T) {
	um, _ := NewTestUserManager()
	bob, _ := um.Create("bob", "bob@example.org", "123456")
	{
		p, err := um.GetProfile(bob)
		assert.Nil(t, err)
		assert.Equal(t, p.UserID, bob.ID)
		assert.Equal(t, p.User.Email, "bob@example.org")
	}
	{
		p, err := um.GetProfile(bob)
		assert.Nil(t, err)
		assert.Equal(t, p.User.Email, "bob@example.org")
	}
	{
		p, _ := um.GetProfile(bob)
		p.Avatar = "mockavator"
		_, err := um.UpdateProfile(bob, p)
		assert.Nil(t, err)
		p, _ = um.GetProfile(bob)
		assert.Equal(t, p.Avatar, "mockavator")
	}
	{
		p, _ := um.GetProfile(bob)
		p.Gender = "male"
		_, err := um.UpdateProfile(bob, p)
		assert.Nil(t, err)
		p, _ = um.GetProfile(bob)
		assert.Equal(t, p.Avatar, "mockavator")
		assert.Equal(t, p.Gender, "male")
	}
}
