package forms

import (
	"net/url"
	"testing"
)

func TestFormError_Get(t *testing.T) {
	form := New(url.Values{})
	form.Required("x")

	isError := form.Errors.Get("x")
	if isError == "" {
		t.Error("Should have an error but did not get one")
	}

	form = New(url.Values{})
	form.Add("x", "value")
	form.Required("x")

	isError = form.Errors.Get("x")
	if isError != "" {
		t.Error("Should not have an error, but got one")
	}

}

func TestForm_Valid(t *testing.T) {
	form := New(url.Values{})

	if !form.Valid() {
		t.Error("Form should have been valid")
	}

	form.Errors.Add("x", "x")

	if form.Valid() {
		t.Error("Form should have been invalid")
	}
}

func TestForm_Required(t *testing.T) {
	form := New(url.Values{})

	form.Required("a", "b")
	if form.Valid() {
		t.Error("Form is valid despite the required fields")
	}

	data := url.Values{}
	data.Add("a", "value")
	data.Add("b", "value")
	form = New(data)

	if !form.Valid() {
		t.Error("Form is still invalid when required fields are valid")
	}
}

func TestForm_Has(t *testing.T) {
	form := New(url.Values{})

	if form.Has("a") {
		t.Error("Form Has() returns true when having a missing field value")
	}

	data := url.Values{}
	data.Add("a", "value")
	form = New(data)

	if !form.Has("a") {
		t.Error("Form Has() returns false despite having the required field value")
	}
}

func TestForm_MinLength(t *testing.T) {
	form := New(url.Values{})

	form.MinLength("a", 2)
	if form.Valid() {
		t.Error("Form min length returns valid for empty field")
	}

	data := url.Values{}
	data.Add("a", "value")
	form = New(data)
	form.MinLength("a", 100)
	if form.Valid() {
		t.Error("Form min length return valid when field is shorter")
	}

	data = url.Values{}
	data.Add("a", "value")
	form = New(data)
	form.MinLength("a", 3)
	if !form.Valid() {
		t.Error("Form min length return invalid when field length is longer")
	}
}

func TestForm_IsEmail(t *testing.T) {
	form := New(url.Values{})

	form.IsEmail("x")
	if form.Valid() {
		t.Error("Form isEmail() shows valid with empty field")
	}

	data := url.Values{}
	data.Add("email", "x@x")
	form = New(data)
	form.IsEmail("email")
	if form.Valid() {
		t.Error("Form isEmail() shows valid with invalid email")
	}

	data = url.Values{}
	data.Add("email", "x@example.com")
	form = New(data)
	form.IsEmail("email")
	if !form.Valid() {
		t.Error("Form isEmail() shows invalid with valid email")
	}
}
