package cache

import (
	"context"
	"github.com/SongOf/edge-storage-core/test"
	"testing"
	"time"
)

func TestRedisCacheLockUnlock(t *testing.T) {
	option := test.RedisOption
	redisCache := NewRedisCache(RedisOption{
		Address:  option["Address"],
		Password: option["Password"],
	})

	ctx := context.Background()
	if err := redisCache.Lock(ctx, "TestLock", "TestLockValue", 10); err != nil {
		t.Error(err)
	}

	client := redisCache.GetClient()

	if val := client.Get(ctx, "TestLock").Val(); val != "TestLockValue" {
		t.Error("set lock failed")
	}

	// lock again will raise error
	if err := redisCache.Lock(ctx, "TestLock", "TestLockValue", 10); err != nil {
		t.Log("lock exists, acquire lock failed")
		t.Log(err)
	}

	// value same test, can't delete different value lock
	if err := redisCache.Unlock(ctx, "TestLock", "TestLockValueTestLockValue"); err != nil {
		t.Log("lock value doesn't match, unlock failed")
		t.Log(err.Error())
	}

	ticker := time.NewTicker(time.Duration(10) * time.Second)
	<-ticker.C
	// 10s later, old lock expired
	if err := redisCache.Lock(ctx, "TestLock", "TestLockValue", 10); err != nil {
		t.Error(err)
	} else {
		t.Log("release expired lock and add new lock")
	}
	defer redisCache.Unlock(ctx, "TestLock", "TestLockValue")
}
