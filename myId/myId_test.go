package myId

import (
	"testing"
)

func TestNextId(t *testing.T) {
	// 测试基本功能
	id1, err := NextId()
	if err != nil {
		t.Fatalf("NextId() error = %v", err)
	}
	if id1 <= 0 {
		t.Errorf("NextId() = %v, want positive", id1)
	}

	// 测试生成多个 ID
	id2, err := NextId()
	if err != nil {
		t.Fatalf("NextId() error = %v", err)
	}
	if id2 <= id1 {
		t.Errorf("NextId() = %v, want > %v", id2, id1)
	}
}

func TestInitSnowFlake(t *testing.T) {
	// 测试初始化
	err := initSnowFlake()
	if err != nil {
		t.Errorf("initSnowFlake() error = %v", err)
	}
}

func BenchmarkNextId(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = NextId()
	}
}
