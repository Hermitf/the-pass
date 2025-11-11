package sms

import (
	"context"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func newTestStore(t *testing.T) (*RedisStore, *miniredis.Miniredis, context.Context) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	t.Cleanup(func() { mr.Close() })

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	store := NewRedisStore(rdb)
	return store, mr, context.Background()
}

func TestRedisStore_SaveGetDeleteCode(t *testing.T) {
	store, mr, ctx := newTestStore(t)
	phone := "13800000000"
	code := "123456"

	// Save
	if err := store.SaveCode(phone, code, 1*time.Second); err != nil {
		t.Fatalf("SaveCode error: %v", err)
	}
	// Get
	got, err := store.GetCode(phone)
	if err != nil {
		t.Fatalf("GetCode error: %v", err)
	}
	if got != code {
		t.Fatalf("GetCode mismatch: got %s want %s", got, code)
	}
	// Expire path: advance miniredis clock beyond TTL
	mr.FastForward(2 * time.Second)
	_, err = store.GetCode(phone)
	if err == nil {
		t.Fatalf("expected error for expired code, got nil")
	}

	// Re-save and delete
	if err := store.SaveCode(phone, code, time.Minute); err != nil {
		t.Fatalf("re-SaveCode error: %v", err)
	}
	if err := store.DeleteCode(phone); err != nil {
		t.Fatalf("DeleteCode error: %v", err)
	}
	_, err = store.GetCode(phone)
	if err == nil {
		t.Fatalf("expected error after delete, got nil")
	}
	_ = ctx // reserved for future context usage
}

func TestRedisStore_CheckRateLimit_and_PeekRate(t *testing.T) {
	store, _, _ := newTestStore(t)
	phone := "13800000001"
	window := 1 * time.Second
	max := 1

	// First check should allow
	allowed, err := store.CheckRateLimit(phone, max, window)
	if err != nil {
		t.Fatalf("CheckRateLimit initial: %v", err)
	}
	if !allowed {
		t.Fatalf("expected first allowed=true")
	}
	// Second within window should block
	allowed, err = store.CheckRateLimit(phone, max, window)
	if err != nil {
		t.Fatalf("CheckRateLimit second: %v", err)
	}
	if allowed {
		t.Fatalf("expected second allowed=false within window")
	}

	// Peek should reflect block and give retryAfter
	peekAllowed, retryAfter, err := store.PeekRate(phone, max, window)
	if err != nil {
		t.Fatalf("PeekRate: %v", err)
	}
	if peekAllowed {
		t.Fatalf("expected peekAllowed=false when rate limited")
	}
	if retryAfter <= 0 || retryAfter > window {
		t.Fatalf("unexpected retryAfter=%s, window=%s", retryAfter, window)
	}

	// Wait for window to pass (ensure we pass the 1s boundary)
	if retryAfter <= 0 {
		retryAfter = window
	}
	time.Sleep(retryAfter + 100*time.Millisecond)

	allowed, err = store.CheckRateLimit(phone, max, window)
	if err != nil {
		t.Fatalf("CheckRateLimit after sleep: %v", err)
	}
	if !allowed {
		t.Fatalf("expected allowed after window passed")
	}
}

func TestRedisStore_DailyCount(t *testing.T) {
	store, _, _ := newTestStore(t)
	phone := "13800000002"

	count, err := store.IncrementDailyCount(phone)
	if err != nil {
		t.Fatalf("IncrementDailyCount(1): %v", err)
	}
	if count != 1 {
		t.Fatalf("count(1)=%d want 1", count)
	}
	count, err = store.IncrementDailyCount(phone)
	if err != nil {
		t.Fatalf("IncrementDailyCount(2): %v", err)
	}
	if count != 2 {
		t.Fatalf("count(2)=%d want 2", count)
	}

	got, ttl, err := store.GetDailyCount(phone)
	if err != nil {
		t.Fatalf("GetDailyCount: %v", err)
	}
	if got != 2 {
		t.Fatalf("GetDailyCount got=%d want 2", got)
	}
	if ttl <= 0 || ttl > 24*time.Hour {
		t.Fatalf("TTL out of range: %s", ttl)
	}
}
