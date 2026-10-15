package myController

import "testing"

func TestParseId(t *testing.T) {
	id, err := ParseId("123")
	if err != nil || id != 123 {
		t.Fatalf("ParseId(123) = (%d, %v)", id, err)
	}
	if _, err := ParseId("abc"); err == nil {
		t.Fatal("ParseId(abc) expected error")
	}
}
