package gorp_test

import (
	"testing"

	"github.com/wcxt/gorp"
)

func TestValidatePath(t *testing.T) {
	tests := []struct {
		path  string
		valid bool
	}{
		{"hello", false},
		{"hello.com", false},
		{"hello.com:80", false},
		{"hello.com/hello", false},
		{"http://", false},
		{"http://hello.com", false},
		{"http://hello.com:80", false},
		{"http://hello.com/hello", false},
		{"/hello/bello", true},
		{"/hello#fragment", false},
		{"/hello?query=value", false},
	}

	for _, test := range tests {
		if err := gorp.ValidatePath(test.path); (err == nil) != test.valid {
			t.Errorf("%s: wanted valid=%v, got error %v", test.path, test.valid, err)
		}
	}
}

func TestValidateUpstream(t *testing.T) {
	tests := []struct {
		upstream string
		valid    bool
	}{
		{"hello", false},
		{"hello.com", false},
		{"hello.com:80", false},
		{"hello.com/hello", false},
		{"http://", false},
		{"http://hello.com/", true},
		{"http://hello.com/?guery=value", false},
		{"http://hello.com/#fragment", false},
		{"http://hello.com", true},
		{"http://hello.com:80", true},
		{"some://hello.com:80", false},
		{"http://hello.com/hello", false},
		{"/hello/bello", false},
	}

	for _, test := range tests {
		if err := gorp.ValidateUpstream(test.upstream); (err == nil) != test.valid {
			t.Errorf("%s: wanted valid=%v, got error %v", test.upstream, test.valid, err)
		}
	}
}
