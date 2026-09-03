package service

// ImportValidationError describes an import file problem that can be shown to
// the user without exposing infrastructure details.
type ImportValidationError struct {
	message string
	cause   error
}

func newImportValidationError(message string, cause error) *ImportValidationError {
	return &ImportValidationError{message: message, cause: cause}
}

func (e *ImportValidationError) Error() string {
	return e.message
}

func (e *ImportValidationError) Unwrap() error {
	return e.cause
}
