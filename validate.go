package gorp

import "fmt"

func ValidatePort(port int) error {
	if port < 0 || port > 65535 {
		return fmt.Errorf("%d: must be in range 0..65535", port)
	}
	return nil
}
