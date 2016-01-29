package main

type HTTPError interface {
	Error() string
	Code() int
}

type httpError struct {
	message string
	code int
}

func NewHTTPError(message string, code int) *httpError {
	return &httpError{
		message: message,
		code: code,
	}
}

func(err *httpError) Error() string {
	return err.message
}

func(err *httpError) Code() int {
	return err.code
}
