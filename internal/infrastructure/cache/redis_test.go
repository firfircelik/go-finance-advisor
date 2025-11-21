package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewRedisCache_Disabled(t *testing.T) {
	cache, err := NewRedisCache("redis://localhost:6379", "", 0, false)
	assert.NoError(t, err)
	assert.NotNil(t, cache)
	assert.False(t, cache.IsEnabled())
}

func TestRedisCache_GetSet_WhenDisabled(t *testing.T) {
	cache := &RedisCache{enabled: false}
	ctx := context.Background()

	// Set should not error when disabled
	err := cache.Set(ctx, "test", "value", 1*time.Minute)
	assert.NoError(t, err)

	// Get should error when disabled
	var result string
	err = cache.Get(ctx, "test", &result)
	assert.Error(t, err)
}
