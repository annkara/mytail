package main

import (
	testing
)

func TestParseArgs(t *testing.T){
	args := {"-n=20", "test"}

	c, err := parseArgs(args)

}