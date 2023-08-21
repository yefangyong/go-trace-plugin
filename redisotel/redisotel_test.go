package redisotel

import (
	"testing"

	"github.com/go-redis/redis/v8"
)

func TestTracingHook(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{})

	if rdb != nil {
		t.Errorf("failed opening connection to redis")
	}

	rdb.AddHook(&TracingHook{})
}
