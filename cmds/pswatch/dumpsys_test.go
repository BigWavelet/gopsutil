package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDumpsys(t *testing.T) {
	pwd, _ := os.Getwd()
	os.Setenv("PATH", os.Getenv("PATH")+":"+filepath.Join(pwd, "testdata"))

	bt := Battery{}
	if err := bt.Update(); err != nil {
		t.Fatal(err)
	}
	if bt.Status != 5 {
		t.Fatalf("Expected status 2, but got %d", bt.Status)
	}
	if bt.Level != 100 {
		t.Fatalf("Expected status 100, but got %d", bt.Level)
	}
	t.Logf("%v", bt)
}
