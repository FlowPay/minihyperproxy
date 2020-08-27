package minihyperproxy

import "errors"

type HttpError struct {
	err  error
	code int
}

var BodyUnmarshallError = &HttpError{err: errors.New("Error unmarshalling body"), code: 422}
var RequestUnmarshallError = &HttpError{err: errors.New("Error unmarshalling request"), code: 422}
var EmptyFieldError = &HttpError{err: errors.New("Required field is empty"), code: 422}
var ServerAlreadyExistsError = &HttpError{err: errors.New("Server with provided name already exists"), code: 500}
