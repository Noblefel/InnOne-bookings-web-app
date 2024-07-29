package forms

import (
	"fmt"
	"net/mail"
	"net/url"
	"strings"
)

// Very basic string form validation
type Form struct {
	url.Values
	Errors map[string][]string
}

func New(data url.Values) *Form {
	return &Form{
		data,
		make(map[string][]string),
	}
}

func (f *Form) Required(fields ...string) {
	for _, field := range fields {
		value := f.Get(field)

		if strings.TrimSpace(value) == "" {
			f.addError(field, "Field cannot be empty")
		}
	}
}

func (f *Form) Has(field string) bool { return len(f.Get(field)) > 0 }

func (f *Form) MinLength(field string, length int) bool {
	if len(f.Get(field)) < length {
		f.addError(field, fmt.Sprintf("Field must be atleast %v characters", length))
		return false
	}

	return true
}

func (f *Form) IsEmail(field string) bool {
	_, err := mail.ParseAddress(f.Get(field))
	if err != nil {
		f.addError(field, "Invalid email address")
		return false
	}

	return true
}

func (f *Form) Valid() bool { return len(f.Errors) == 0 }

func (f *Form) addError(field, msg string) {
	f.Errors[field] = append(f.Errors[field], msg)
}
