package main

import (
	//"regexp/syntax"
	"strings"
	"testing"
)

func TestCapitalization(t *testing.T) {
	r1 := strings.Title("foo_bar_baz")
	if r1 != "Foo_bar_baz" {
		t.Errorf("Capitalization doesnt match: %v", r1)
	}

	r2 := strings.Title("foo bar baz - 2011")
	if r2 != "Foo Bar Baz - 2011" {
		t.Errorf("Capitalization doesnt match: %v", r2)
	}
}

/*
func TestRegtexReplace(t *testing.T) {
	re, err := syntax.Parse("s/_/\\ ", 0)
	if err != nil {
		t.Fatalf("Failed to parse regex: %v", err)
	}

	t.Error("Replacement doesnt match")
}
*/
