package main

import (
	"bytes"
	"testing"
)

func TestInputIntervalMilliSecond(t *testing.T) {
	in := bytes.NewBufferString("1000")
	got := inputIntervalMilliSecond(in)
	want := 1000
	if got != want {
		t.Errorf("got %d, want %d", got, want)
	}
}
