package web

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

const (
	CodeValidation   = "VALIDATION_ERROR"
	CodeUnauthorized = "UNAUTHORIZED"
	CodeForbidden    = "FORBIDDEN"
	CodeNotFound     = "NOT_FOUND"
	CodeConflict     = "CONFLICT"
	CodeBadGateway   = "BAD_GATEWAY"
	CodeUnavailable  = "UNAVAILABLE"
	CodeInternal     = "INTERNAL_ERROR"
)

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type errorBody struct {
	Code    string       `json:"code"`
	Message string       `json:"message"`
	Details []FieldError `json:"details,omitempty"`
}

type errorResponse struct {
	Error errorBody `json:"error"`
}

func WriteJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if payload == nil {
		return
	}
	_ = json.NewEncoder(w).Encode(payload)
}

func WriteError(w http.ResponseWriter, status int, code, message string) {
	WriteJSON(w, status, errorResponse{Error: errorBody{Code: code, Message: message}})
}

func WriteValidation(w http.ResponseWriter, details []FieldError) {
	WriteJSON(w, http.StatusBadRequest, errorResponse{
		Error: errorBody{
			Code:    CodeValidation,
			Message: "Request validation failed",
			Details: details,
		},
	})
}

func DecodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		WriteError(w, http.StatusBadRequest, CodeValidation, "Malformed JSON body")
		return false
	}
	return true
}

func ParsePage(r *http.Request) (page, pageSize int, details []FieldError) {
	page = 1
	pageSize = 20

	if raw := strings.TrimSpace(r.URL.Query().Get("page")); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 1 {
			details = append(details, FieldError{Field: "page", Message: "must be an integer >= 1"})
		} else {
			page = v
		}
	}
	if raw := strings.TrimSpace(r.URL.Query().Get("pageSize")); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 1 || v > 100 {
			details = append(details, FieldError{Field: "pageSize", Message: "must be an integer between 1 and 100"})
		} else {
			pageSize = v
		}
	}
	return page, pageSize, details
}

type Page[T any] struct {
	Items    []T `json:"items"`
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
	Total    int `json:"total"`
}
