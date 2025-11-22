package domain

import (
	"errors"
	"fmt"
)

type ErrorCode string

const (
	ErrTeamExists  ErrorCode = "TEAM_EXISTS"
	ErrPRExists    ErrorCode = "PR_EXISTS"
	ErrPRMerged    ErrorCode = "PR_MERGED"
	ErrNotAssigned ErrorCode = "NOT_ASSIGNED"
	ErrNoCandidate ErrorCode = "NO_CANDIDATE"
	ErrNotFound    ErrorCode = "NOT_FOUND"
)

type DomainError struct {
	Code    ErrorCode
	Message string
	Err     error
}

func (e *DomainError) Error() string {
	if e.Err == nil {
		return fmt.Sprintf("%s: %s", e.Code, e.Message)
	}
	return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
}

func (e *DomainError) Unwrap() error {
	return e.Err
}

func NewDomainError(code ErrorCode, msg string) *DomainError {
	return &DomainError{
		Code:    code,
		Message: msg,
	}
}

func WrapDomainError(code ErrorCode, msg string, err error) *DomainError {
	return &DomainError{
		Code:    code,
		Message: msg,
		Err:     err,
	}
}

func IsDomainError(err error, code ErrorCode) bool {
	var de *DomainError
	if !errors.As(err, &de) {
		return false
	}
	return de.Code == code
}

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	if e.Field == "" {
		return fmt.Sprintf("validation error: %s", e.Message)
	}
	return fmt.Sprintf("validation error: field %q: %s", e.Field, e.Message)
}

func NewValidationError(field, msg string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: msg,
	}
}
