package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsNil(t *testing.T) {
	t.Run("int any", func(t *testing.T) {
		var i any = 10
		assert.False(t, isNil(i))
	})

	t.Run("nil map any", func(t *testing.T) {
		var m any = map[string]int(nil)
		fmt.Println(m)
		assert.True(t, isNil(m))
	})
}
