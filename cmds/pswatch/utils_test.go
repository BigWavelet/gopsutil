package main

import "testing"

func TestUniqSlice(t *testing.T) {
	out := UniqSlice([]string{"yy", "ss", "yy"})
	if len(out) != 2 {
		t.Fatal("Only 2 string should be left")
	}
}
