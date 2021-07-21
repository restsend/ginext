package ginext

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSignals(t *testing.T) {
	{
		s := NewSignals()
		r := false
		s.Connect("hello", func(sender interface{}, params ...interface{}) {
			r = true
		})
		s.Emit("hello", nil)
		assert.True(t, r)
	}
	{
		s := NewSignals()
		s.Emit("hello", nil)
	}
	{
		s := NewSignals()
		r := 0
		s.Connect("hello", func(sender interface{}, params ...interface{}) {
			r += 1
			s.Connect("hello", func(sender interface{}, params ...interface{}) {
				r += 1
			})
		})
		s.Emit("hello", nil)
		assert.Equal(t, 1, r)
		s.Emit("hello", nil)
		assert.Equal(t, 3, r)
	}
	{
		s := NewSignals()
		var sid uint
		sid = s.Connect("hello", func(sender interface{}, params ...interface{}) {
			s.Disconnect("hello", sid)
		})
		s.Emit("hello", nil)
		assert.Equal(t, 0, len(s.sigHandlers["hello"]))
	}
}
