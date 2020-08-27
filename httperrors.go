package minihyperproxy

type HttpError struct {
	ErrString string `json:"Error"`
	code      int
}

func (h *HttpError) Error() string {
	return h.ErrString
}

var BodyUnmarshallError = &HttpError{ErrString: "Error unmarshalling body", code: 422}
var RequestUnmarshallError = &HttpError{ErrString: "Error unmarshalling request", code: 422}
var EmptyFieldError = &HttpError{ErrString: "Required field is empty", code: 422}
var ServerAlreadyExistsError = &HttpError{ErrString: "Server with provided name already exists", code: 500}
var NoServerFoundError = &HttpError{ErrString: "Server Not Found", code: 500}
var WrongServerTypeError = &HttpError{ErrString: "Wrong Server Type", code: 500}
