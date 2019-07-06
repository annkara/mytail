package main

import "testing"

func TestParseArgs(t *testing.T) {
	args := []string{"-n=20", "test"}

	c, err := parseArgs(args)
	if err != nil {
		t.Errorf("failed:")
	}
	if c == nil {
		t.Error()
	}
}
