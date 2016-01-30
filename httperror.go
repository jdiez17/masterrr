package main

import (
	"encoding/json"
	"net/http"
)

type HTTPError interface {
	Error() string
	Code() int
	Write(http.ResponseWriter)
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

func(err *httpError) Write(w http.ResponseWriter) {
	out, _ := json.Marshal(StatusMessageResponse{
		Message: err.Error(),
		Success: false,
	})

	w.WriteHeader(err.Code())
	w.Write(out)
}
