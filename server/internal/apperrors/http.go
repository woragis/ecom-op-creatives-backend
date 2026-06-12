package apperrors

import (
	"encoding/json"
	"errors"
	"net/http"
)

type errorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

const CodeInternal = "INTERNAL_ERROR"

func HTTPStatus(err error) int {
	if err == nil {
		return http.StatusOK
	}
	if ae, ok := As(err); ok {
		switch ae.Kind {
		case KindInvalid:
			return http.StatusBadRequest
		case KindUnauthorized:
			return http.StatusUnauthorized
		case KindForbidden:
			return http.StatusForbidden
		case KindNotFound:
			return http.StatusNotFound
		case KindConflict:
			return http.StatusConflict
		case KindTooManyRequests:
			return http.StatusTooManyRequests
		case KindUnavailable:
			return http.StatusServiceUnavailable
		default:
			return http.StatusInternalServerError
		}
	}
	return http.StatusInternalServerError
}

func PublicPayload(err error) (code string, message string) {
	if err == nil {
		return "", ""
	}
	if ae, ok := As(err); ok {
		return ae.Code, ae.Message
	}
	return CodeInternal, "An unexpected error occurred."
}

func WriteError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}
	code, msg := PublicPayload(err)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(HTTPStatus(err))
	_ = json.NewEncoder(w).Encode(errorBody{Code: code, Message: msg})
}

func IsNotFound(err error) bool {
	var ae *Error
	if !errors.As(err, &ae) || ae == nil {
		return false
	}
	return ae.Kind == KindNotFound
}
