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

func compareArrays(arr1, arr2 []string) bool {
	if len(arr1) != len(arr2) {
		return false
	}

	for i, _ := range arr1 {
		if arr1[i] != arr2[i] {
			return false
		}
	}

	return true
}
func TestChunked(t *testing.T) {
	data1 := []string{"0", "1", "2", "3", "4"}
	chunked1a := chunked(data1, 2)
	if len(chunked1a) != 3 {
		t.Errorf("Incorrect amount of chunks")
	}
	if !compareArrays(chunked1a[0], []string{"0", "1"}) {
		t.Errorf("Chunk contents incorrect")
	}
	if !compareArrays(chunked1a[1], []string{"2", "3"}) {
		t.Errorf("Chunk contents incorrect")
	}
	if !compareArrays(chunked1a[2], []string{"4"}) {
		t.Errorf("Chunk contents incorrect")
	}

	chunked1b := chunked(data1, 4)
	if len(chunked1b) != 2 {
		t.Errorf("Incorrect amount of chunks")
	}
	if !compareArrays(chunked1b[0], []string{"0", "1", "2", "3"}) {
		t.Errorf("Chunk contents incorrect")
	}
	if !compareArrays(chunked1b[1], []string{"4"}) {
		t.Errorf("Chunk contents incorrect")
	}
}
