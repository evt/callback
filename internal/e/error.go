package e

import (
	"fmt"
	"net/http"
)

type Error interface {
	Code() int
	Detail() string
}

type httpError struct {
	detail string
	code   int
}

func (e httpError) Error() string {
	return fmt.Sprintf(`code: %d, detail: '%s'`, e.code, e.detail)
}

func (e httpError) Code() int {
	return e.code
}

func (e httpError) Detail() string {
	return e.detail
}

func NewInternal(detail string) Error {
	return httpError{
		detail: detail,
		code:   http.StatusInternalServerError,
	}
}

func NewInternalf(template string, args ...interface{}) Error {
	return httpError{
		detail: fmt.Sprintf(template, args...),
		code:   http.StatusInternalServerError,
	}
}

func NewNotFound(detail string) Error {
	return httpError{
		detail: detail,
		code:   http.StatusNotFound,
	}
}

func NewBadRequest(detail string) Error {
	return httpError{
		detail: detail,
		code:   http.StatusBadRequest,
	}
}
