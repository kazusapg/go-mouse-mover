package main

import (
	"bytes"
	"context"
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

func TestMoveMousePosition(t *testing.T) {
	t.Run("Check to see if the context can be used to cancel successfully", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		mi := moveInfo{Positions: []position{{X: 0, Y: 0}}, IntervalMillisecond: 10000}
		go moveMousePosition(ctx, mi)
		cancel()
	})
}
