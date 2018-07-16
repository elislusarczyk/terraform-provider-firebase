package firebase

import (
	"fmt"
	"net/mail"
	"net/url"
	"regexp"
)

func validateEmail(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)

	if _, err := mail.ParseAddress(value); err != nil {
		errors = append(errors, fmt.Errorf(
			"%q should be an email: %q",
			k, value))
	}
	return
}

func validateURL(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)

	if _, err := url.ParseRequestURI(value); err != nil {
		errors = append(errors, fmt.Errorf(
			"%q should be an URL: %q",
			k, value))
	}
	return
}

func validateE164PhoneNumber(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)

	// https://en.wikipedia.org/wiki/E.164
	if !regexp.MustCompile(`^\+(?:[0-9]●?){6,14}[0-9]$`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"%q should be a E.164: %q",
			k, value))
	}
	return
}
