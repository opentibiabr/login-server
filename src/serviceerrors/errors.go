package serviceerrors

import (
	"errors"
	"fmt"
)

const (
	// Error code ranges:
	// 1-99: global login/authentication errors.
	// 2000-2999: database-backed login data and session persistence errors.
	// 3000-3999: login service operational errors.
	// 4000-4999: game data/configuration availability errors.
	// New codes should use the next free value in the matching range.
	CodeInvalidCredentials        = 3
	CodeDatabaseUnavailable       = 2001
	CodeAccountDataUnavailable    = 2002
	CodeCharacterListLoadFailed   = 2003
	CodeSessionStorageUnavailable = 2004
	CodeSessionCreateFailed       = 2005
	CodeLoginServiceUnavailable   = 3001
	CodeUnsupportedRequestType    = 3002
	CodeEventScheduleUnavailable  = 4001
	CodeBoostedDataUnavailable    = 4002
)

const invalidCredentialsMessage = "Account email or password is not correct."

type PublicError struct {
	Code    int
	Name    string
	Message string
	Hint    string
	Cause   error
}

func New(code int, name string, message string, cause error) *PublicError {
	return &PublicError{
		Code:    code,
		Name:    name,
		Message: message,
		Cause:   cause,
	}
}

func WithHint(err *PublicError, hint string) *PublicError {
	if err == nil || hint == "" {
		return err
	}

	copied := *err
	copied.Hint = hint
	return &copied
}

func MessageWithHint(err *PublicError) string {
	if err == nil {
		return ""
	}
	if err.Hint == "" {
		return err.Message
	}
	return fmt.Sprintf("%s Admin hint: %s", err.Message, err.Hint)
}

func InvalidCredentials() *PublicError {
	return New(CodeInvalidCredentials, "INVALID_CREDENTIALS", invalidCredentialsMessage, nil)
}

func LoginService(code int, name string, cause error) *PublicError {
	return New(code, name, fmt.Sprintf("Login service error. Please contact support. Error: %s (LS-%d).", name, code), cause)
}

func GameData(code int, name string, cause error) *PublicError {
	return New(code, name, fmt.Sprintf("Game data configuration error. Please contact support. Error: %s (LS-%d).", name, code), cause)
}

func (err *PublicError) Error() string {
	if err == nil {
		return ""
	}

	if err.Cause == nil {
		return fmt.Sprintf("%s (LS-%d): %s", err.Name, err.Code, err.Message)
	}

	return fmt.Sprintf("%s (LS-%d): %s: %v", err.Name, err.Code, err.Message, err.Cause)
}

func (err *PublicError) Unwrap() error {
	if err == nil {
		return nil
	}

	return err.Cause
}

func FromError(err error) (*PublicError, bool) {
	var publicErr *PublicError
	if errors.As(err, &publicErr) {
		return publicErr, true
	}

	return nil, false
}
