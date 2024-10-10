package gorp

import (
	"errors"
	"net/url"
)

func ValidatePort(port int) error {
	if port < 0 || port > 65535 {
		return errors.New("must be in range 0..65535")
	}
	return nil
}

func ValidatePath(path string) error {
	parsed, err := url.ParseRequestURI(path)
	if err != nil {
		return errors.New("must be a valid URI")
	}

	// ParseRequestURI ignores fragment
	parsed, err = url.Parse(path)
	if err != nil {
		panic(err)
	}

	if parsed.Scheme != "" || parsed.Host != "" {
		err = errors.New("must be a relative URI")
	}
	if parsed.RawQuery != "" {
		err = errors.Join(err, errors.New("must not include query parameters"))
	}
	if parsed.Fragment != "" {
		err = errors.Join(err, errors.New("must not include fragment identifier"))
	}

	return err
}

func ValidateUpstream(upstream string) error {
	parsed, err := url.Parse(upstream)
	if err != nil {
		return errors.New("must be a valid URL")
	}

	if parsed.Scheme == "" || parsed.Host == "" {
		return errors.New("must be a absolute URL")
	}
	if parsed.Scheme != "http" {
		err = errors.New("unsupported URL scheme")
	}
	if parsed.Path != "" && parsed.Path != "/" {
		err = errors.Join(err, errors.New("must have root path"))
	}
	if parsed.RawQuery != "" {
		err = errors.Join(err, errors.New("must not include query parameters"))
	}
	if parsed.Fragment != "" {
		err = errors.Join(err, errors.New("must not include fragment identifier"))
	}

	return err
}
