package main

import "testing"

func TestDumpsys(t *testing.T) {
	bt := Battery{}
	if err := bt.Update(); err != nil {
		t.Fatal(err)
	}
	t.Logf("%v", bt)
}
