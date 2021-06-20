package ginext

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQuote(t *testing.T) {
	um, _ := NewTestUserManager()
	var testUser GinExtUser
	um.db.AutoMigrate(&GinExtUser{})
	result := um.db.Where(&GinExtUser{
		UserName: "bob",
	}).Take(&testUser)

	assert.Error(t, result.Error)

	var bob GinExtUser
	bob.UserName = "bob"
	bob.Enabled = true
	um.db.Create(&bob)
	vals := map[string]interface{}{
		"Email":   "bob@example.org",
		"Enabled": nil,
	}
	result = um.db.Model(&GinExtUser{}).Where("ID", bob.ID).UpdateColumns(vals)
	assert.Nil(t, result.Error)
	result = um.db.Where("id", bob.ID).Take(&testUser)
	assert.Nil(t, result.Error)
	assert.False(t, testUser.Enabled)

}

func TestGinConfig(t *testing.T) {
	cfg := NewGinExt("..")
	cfg.Init()
	v := cfg.GetValue("Hello")
	assert.Equal(t, v, "")

	cfg.SetValue("Hello", "12345")
	v = cfg.GetValue("Hello")
	assert.Equal(t, v, "12345")

	v = cfg.GetValue("HELLO")
	assert.Equal(t, v, "12345")
}

func TestGinUniqueKey(t *testing.T) {
	cfg := NewGinExt("..")
	cfg.Init()
	tx := cfg.DbInstance.Model(&GinExtUser{})
	cfg.DbInstance.AutoMigrate(&GinExtUser{})
	key0 := GenUniqueKey(tx, "user_name", 4)
	u0 := GinExtUser{
		ID:       0,
		UserName: key0,
	}
	result := cfg.DbInstance.Create(&u0)
	assert.Nil(t, result.Error)
	result = cfg.DbInstance.Create(&u0)
	assert.NotNil(t, result.Error)

}
