package lib

import (
	"testing"
)

func TestHostNameFromHostPortReturnsExpected(t *testing.T) {
	given := "example.com:8080"
	expected := "example.com"

	result := hostNameFromHostPort(given)
	if result != expected {
		t.Errorf("Expected `%s`, but got `%s`", expected, result)
	}
}

func TestHostNameFromHostPortReturnHostnameWhenNoPortGiven(t *testing.T) {
	given := "example.com"
	expected := "example.com"

	result := hostNameFromHostPort(given)
	if result != expected {
		t.Errorf("Expected `%s`, but got `%s`", expected, result)
	}
}

func TestHostNameFromHostPortReturnsBlankWhenEmpty(t *testing.T) {
	given := ""
	expected := ""

	result := hostNameFromHostPort(given)
	if result != expected {
		t.Errorf("Expected `%s`, but got `%s`", expected, result)
	}
}
