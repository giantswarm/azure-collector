package setting

import "fmt"

type Azure struct {
	EnvironmentName string
	Location        string
}

func (a Azure) Validate() error {
	if a.EnvironmentName == "" {
		return fmt.Errorf("EnvironmentName must not be empty")
	}
	if a.Location == "" {
		return fmt.Errorf("location must not be empty")
	}

	return nil
}
