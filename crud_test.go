package ginext

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockObj struct {
	ID      uint `gorm:"primarykey"`
	Title   string
	Remark  string
	Deleted bool
}

func TestCRUD(t *testing.T) {
	ext := NewGinExt("..")
	ext.Init()
	ext.DbInstance.AutoMigrate(&mockObj{})
	vals := map[string]interface{}{
		"Title":  "xxx",
		"Remark": "test",
	}
	u := mockObj{
		Title:  "yyy",
		Remark: "1234",
	}
	NewObject(nil, ext.DbInstance, &u)
	vals = map[string]interface{}{}
	EditObject(nil, ext.DbInstance, &u, u.ID, vals)
	ext.DbInstance.First(&u)
	assert.Equal(t, "yyy", u.Title)
	assert.Equal(t, "1234", u.Remark)

	DeleteObject(nil, ext.DbInstance, &u, u.ID, true)

	ext.DbInstance.Model(&u).First(&u)
	assert.True(t, u.Deleted)
	DeleteObject(nil, ext.DbInstance, &u, u.ID, false)
	r := ext.DbInstance.Model(&u).First(&u)
	assert.NotNil(t, r.Error)
}
