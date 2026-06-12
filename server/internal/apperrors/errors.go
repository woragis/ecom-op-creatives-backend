package apperrors

import (
	"errors"
	"fmt"
)

type Error struct {
	Code    string
	Message string
	Kind    Kind
	Cause   error
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}

func (e *Error) Unwrap() error { return e.Cause }

func invalid(code, msg string, cause error) *Error {
	return &Error{Code: code, Message: msg, Kind: KindInvalid, Cause: cause}
}

func notFound(code, msg string, cause error) *Error {
	return &Error{Code: code, Message: msg, Kind: KindNotFound, Cause: cause}
}

func conflict(code, msg string, cause error) *Error {
	return &Error{Code: code, Message: msg, Kind: KindConflict, Cause: cause}
}

func internal(code, msg string, cause error) *Error {
	return &Error{Code: code, Message: msg, Kind: KindInternal, Cause: cause}
}

func unavailable(code, msg string, cause error) *Error {
	return &Error{Code: code, Message: msg, Kind: KindUnavailable, Cause: cause}
}

func Invalid(code, msg string) *Error                         { return invalid(code, msg, nil) }
func InvalidCause(code, msg string, cause error) *Error       { return invalid(code, msg, cause) }
func NotFound(code, msg string) *Error                        { return notFound(code, msg, nil) }
func NotFoundCause(code, msg string, cause error) *Error      { return notFound(code, msg, cause) }
func ConflictErr(code, msg string) *Error                     { return conflict(code, msg, nil) }
func InternalErr(code, msg string) *Error                     { return internal(code, msg, nil) }
func InternalCause(code, msg string, cause error) *Error      { return internal(code, msg, cause) }
func Unavailable(code, msg string) *Error                     { return unavailable(code, msg, nil) }
func UnavailableCause(code, msg string, cause error) *Error   { return unavailable(code, msg, cause) }

func Wrapf(cause error, format string, args ...any) *Error {
	if cause == nil {
		return nil
	}
	return internal(CodeInternal, "An unexpected error occurred.", fmt.Errorf("%s: %w", fmt.Sprintf(format, args...), cause))
}

func As(err error) (*Error, bool) {
	var ae *Error
	if errors.As(err, &ae) {
		return ae, true
	}
	return nil, false
}
