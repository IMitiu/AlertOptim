package main

import (
	"os"
	"testing"
)

func TestSimple(t *testing.T) {
	a, errNew := NewSimple("source.yaml")
	require.Nil(t, errNew)

	a.Spool(os.Stdout)
}
