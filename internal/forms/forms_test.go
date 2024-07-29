package forms

import (
	"net/url"
	"testing"
)

func TestForm_Valid(t *testing.T) {
	form := New(url.Values{})

	if !form.Valid() {
		t.Error("should have been valid")
	}

	form.addError("x", "x")

	if form.Valid() {
		t.Error("should have been invalid")
	}
}

func TestForm_Required(t *testing.T) {
	form := New(url.Values{})

	form.Required("a", "b")
	if form.Valid() {
		t.Error("valid despite the required fields")
	}

	data := url.Values{"a": {"value"}, "b": {"value"}}
	form = New(data)

	if !form.Valid() {
		t.Error("invalid when required fields are valid")
	}
}

func TestForm_Has(t *testing.T) {
	form := New(url.Values{})

	if form.Has("a") {
		t.Error("returns true when having a missing field value")
	}

	data := url.Values{"a": {"value"}}
	form = New(data)

	if !form.Has("a") {
		t.Error("returns false despite having the required field value")
	}
}

func TestForm_MinLength(t *testing.T) {
	form := New(url.Values{})

	form.MinLength("a", 2)
	if form.Valid() {
		t.Error("returns valid for empty field")
	}

	form = New(url.Values{"a": {"value"}})
	form.MinLength("a", 100)
	if form.Valid() {
		t.Error("return valid when field is shorter")
	}

	form = New(url.Values{"a": {"value"}})
	form.MinLength("a", 3)
	if !form.Valid() {
		t.Error("return invalid when field length is longer")
	}
}

func TestForm_IsEmail(t *testing.T) {
	form := New(url.Values{"email": {"a@."}})
	form.IsEmail("email")
	if form.Valid() {
		t.Error("returns valid with invalid email")
	}

	form = New(url.Values{"email": {"abcdefg@gmail.com"}})
	form.IsEmail("email")
	if !form.Valid() {
		t.Error("returns invalid with valid email")
	}
}
