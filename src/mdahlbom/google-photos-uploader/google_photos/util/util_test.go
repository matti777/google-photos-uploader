package util

import (
	"bytes"
	"io/ioutil"
	"testing"
)

var Fnc func(string, ...interface{}) = nil

func TestCreateImageUploadRequest(t *testing.T) {
	uri := "http://test.com/path"
	mimeType := "application/foobar"
	bodyString := "ABC"
	data := []byte(bodyString)

	r, err := NewImageUploadRequest(uri, mimeType, bytes.NewReader(data), nil)
	if err != nil {
		t.Errorf("Failed to create upload request: %v", err)
	}

	if r == nil {
		t.Fatalf("Got nil request")
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		t.Errorf("Reading request body failed: %v", err)
	}

	if string(body) != bodyString {
		t.Errorf("Invalid request body")
	}
}
