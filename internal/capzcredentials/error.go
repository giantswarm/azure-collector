package capzcredentials

import "github.com/giantswarm/microerror"

var credentialsNotFoundError = &microerror.Error{
	Kind: "credentialsNotFoundError",
}

// IsCredentialsNotFoundError asserts credentialsNotFoundError.
func IsCredentialsNotFoundError(err error) bool {
	return microerror.Cause(err) == credentialsNotFoundError
}

var invalidObjectMetaError = &microerror.Error{
	Kind: "invalidObjectMetaError",
}

var missingIdentityRefError = &microerror.Error{
	Kind: "missingIdentityRefError",
}

// IsMissingIdentityRef asserts missingIdentityRefError.
func IsMissingIdentityRef(err error) bool {
	return microerror.Cause(err) == missingIdentityRefError
}

var missingValueError = &microerror.Error{
	Kind: "missingValueError",
}

var tooManyCredentialsError = &microerror.Error{
	Kind: "tooManyCredentialsError",
}

// IsTooManyCredentials asserts tooManyCredentialsError.
func IsTooManyCredentials(err error) bool {
	return microerror.Cause(err) == tooManyCredentialsError
}
