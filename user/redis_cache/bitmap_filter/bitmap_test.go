package bitmap_filter

import (
	"github.com/redis/go-redis/v9"
	"testing"
)

func TestBitmapFilter(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	// stablished
	bf := NewBitMapFilterDefault(rdb, "test_bitmap_filter", 1000)

	// Test adding elements
	err := bf.Add("element1")
	if err != nil {
		t.Errorf("Failed to add element1: %v", err)
		return
	}

	// Test checking existence
	if !bf.Exist("element1") {
		t.Error("Expected element1 to exist")
	}
	if bf.Exist("element3") {
		t.Error("Expected element3 to not exist")
	}

}
