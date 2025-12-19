package connection_history

import "errors"

// ErrPasswordNotFound is returned when a password is not found in the keyring
var ErrPasswordNotFound = errors.New("password not found in keyring")

// PasswordSaveError represents an error that occurred while saving a password
type PasswordSaveError struct {
	Err     error
	Message string
}

func (e *PasswordSaveError) Error() string {
	if e.Message != "" {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Err.Error()
}

func (e *PasswordSaveError) Unwrap() error {
	return e.Err
}

// PasswordReadError represents an error that occurred while reading a password
type PasswordReadError struct {
	Err error
}

func (e *PasswordReadError) Error() string {
	return "failed to read password from keyring: " + e.Err.Error()
}

func (e *PasswordReadError) Unwrap() error {
	return e.Err
}
