package gorp

import (
	"errors"
	"fmt"
	"net/url"
)

func ValidatePort(port int) error {
	if port < 0 || port > 65535 {
		return fmt.Errorf("%d: must be in range 0..65535", port)
	}
	return nil
}

func ValidateSourceFlag(source string) error {
	parsed, err := url.ParseRequestURI(source)
	if err != nil {
		return fmt.Errorf("%s: must be a valid URI", source)
	}

	// ParseRequestURI ignores fragment
	parsed, _ = url.Parse(source)

	if parsed.Scheme != "" || parsed.Host != "" {
		err = fmt.Errorf("%s: must be a relative URI", source)
	}
	if parsed.RawQuery != "" {
		err = errors.Join(err, fmt.Errorf("%s: must not include query parameters", source))
	}
	if parsed.Fragment != "" {
		err = errors.Join(err, fmt.Errorf("%s: must not include fragment identifier", source))
	}

	return err
}
