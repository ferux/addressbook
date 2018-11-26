package controllers

// InternalError describes error taht was caused by bad entries.
type InternalError struct {
	msg string
}

func (err *InternalError) Error() string {
	return err.msg
}

// NewInternalError creates new InternalError error.
func NewInternalError(msg string) error {
	return &InternalError{msg: msg}
}
