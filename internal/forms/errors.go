package forms

type errors map[string][]string

// Add an error message to a given form field
func (e errors) Add(field, message string) {
	e[field] = append(e[field], message)
}

// Get returns the first error message to a given form field
func (e errors) Get(field string) string {
	str := e[field]
	if len(str) == 0 {
		return ""
	}

	return str[0]
}
