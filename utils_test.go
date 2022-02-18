package ginext

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormat(t *testing.T) {
	fmtText := "{{.Text}}"
	ctx := map[string]string{
		"Text": "Hello",
	}

	fallBack := "{{.Text}}"
	{
		v := FormatData(fmtText, ctx, fallBack)
		assert.Equal(t, "Hello", v)
	}
	{
		v := FormatData("{{.+}}", ctx, fallBack)
		assert.Equal(t, "Hello", v)
	}
}
func TestSafeCall(t *testing.T) {
	SafeCall(func() error {
		return nil
	}, nil)

	SafeCall(func() error {
		panic("mock")
	}, func(e error) {
		assert.Equal(t, "mock", e.Error())
	})
}

type testMapForm struct {
	ID     uint    `json:"id" binding:"required"`
	Title  *string `json:"title"`
	Source *string `json:"source"`
}

func TestFormAsMap(t *testing.T) {
	title := "title"
	form := testMapForm{
		ID:    100,
		Title: &title,
	}
	{
		vals := FormAsMap(form, []string{"Title", "Target"})
		assert.Equal(t, 1, len(vals))
		assert.Equal(t, title, vals["Title"])
	}
	{
		vals := FormAsMap(form, []string{"ID", "Source"})
		assert.Equal(t, 1, len(vals))
		assert.Equal(t, uint(100), vals["ID"])
	}
	{
		vals := FormAsMap(&form, []string{"ID", "Source"})
		assert.Equal(t, 1, len(vals))
		assert.Equal(t, uint(100), vals["ID"])
	}
}
