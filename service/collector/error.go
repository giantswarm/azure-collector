package collector

import (
	"net/http"

	"github.com/Azure/go-autorest/autorest"
	"github.com/giantswarm/microerror"
)

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

// IsThrottlingError asserts 429 response.
func IsThrottlingError(err error) bool {
	if err == nil {
		return false
	}

	c := microerror.Cause(err)

	{
		dErr, ok := c.(autorest.DetailedError)
		if ok {
			if dErr.StatusCode == http.StatusTooManyRequests {
				return true
			}
		}
	}

	return false
}
