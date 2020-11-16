package contactcache

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestRedisCache(t *testing.T) {
	//Spin up local test redis
	s, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	defer s.Close()

	//Set config to test redis
	viper.Set("cache.address", s.Addr())

	cache, err := NewRedisCache()
	if assert.NoError(t, err) {
		ctx := context.Background()

		key := "foo"

		//Empty state
		resp, err := cache.Get(ctx, key)
		assert.Error(t, err, "expected error from redis on empty")
		assert.Equal(t, "", resp)

		//Set key
		err = cache.Set(ctx, key, "bar", 1*time.Minute)
		if err != nil {
			t.Fatal(err)
		}

		//Get key again
		resp, err = cache.Get(ctx, key)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "bar", resp)

		err = cache.Delete(ctx, key)
		if err != nil {
			t.Fatal(err)
		}

		//Delete prefixes
		cache.Set(ctx, key, "bar", 1*time.Minute)
		err = cache.Delete(ctx, "f*")
		if err != nil {
			t.Fatal(err)
		}
		resp, _ = cache.Get(ctx, key)
		assert.Equal(t, "", resp)
	}
}
