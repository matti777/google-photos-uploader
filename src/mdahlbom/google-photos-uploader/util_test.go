package main

import (
	"strings"
	"testing"
)

func TestReplaceInString(t *testing.T) {
	if r, err := replaceInString("foo_bar", "_, ,\"'\",-"); err != nil {
		t.Fatalf("Failed to replace: %v", err)
	} else if r != "foo bar" {
		t.Errorf("Replacement incorrect: %v", r)
	}

	if r, err := replaceInString("foo_bar_baz-2010", "_, ,-, - "); err != nil {
		t.Fatalf("Failed to replace: %v", err)
	} else if r != "foo bar baz - 2010" {
		t.Errorf("Replacement incorrect: %v", r)
	}

	if r, err := replaceInString("foo_bar_baz", ""); err != nil {
		t.Fatalf("Failed to replace: %v", err)
	} else if r != "foo_bar_baz" {
		t.Errorf("Replacement incorrect: %v", r)
	}
}

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
