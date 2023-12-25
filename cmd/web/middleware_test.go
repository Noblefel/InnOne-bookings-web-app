package main

import (
	"net/http"
	"testing"
)

func TestCSRF(t *testing.T) {
	var myH myHandler

	h := CSRF(&myH)

	switch v := h.(type) {
	case http.Handler:
		// Passed
	default:
		t.Errorf("Type is not http.Handler, but %T", v)
	}
}

func TestSessionLoad(t *testing.T) {
	var myH myHandler

	h := SessionLoad(&myH)

	switch v := h.(type) {
	case http.Handler:
		// Passed
	default:
		t.Errorf("Type is not http.Handler, but %T", v)
	}
}
