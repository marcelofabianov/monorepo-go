package web

import (
	"encoding/json"
	"net/http"

	"github.com/marcelofabianov/fault"
)

type ErrorResponse struct {
	Code       string `json:"code" example:"VALIDATION_ERROR"`
	Message    string `json:"message" example:"Invalid request parameters"`
	StatusCode int    `json:"status_code" example:"400"`
}

type MessageResponse struct {
	Message string `json:"message" example:"Operation completed successfully"`
}

func Success(w http.ResponseWriter, r *http.Request, status int, data any) {
	writeJSON(w, status, data)
}

func Error(w http.ResponseWriter, r *http.Request, err error) {
	response := fault.ToResponse(err)
	writeJSON(w, response.StatusCode, response)
}

func Created(w http.ResponseWriter, r *http.Request, data any) {
	Success(w, r, http.StatusCreated, data)
}

func NoContent(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func Accepted(w http.ResponseWriter, r *http.Request, data any) {
	Success(w, r, http.StatusAccepted, data)
}

func BadRequest(w http.ResponseWriter, r *http.Request, err error) {
	Error(w, r, err)
}

func Unauthorized(w http.ResponseWriter, r *http.Request, err error) {
	Error(w, r, err)
}

func Forbidden(w http.ResponseWriter, r *http.Request, err error) {
	Error(w, r, err)
}

func NotFound(w http.ResponseWriter, r *http.Request, err error) {
	Error(w, r, err)
}

func Conflict(w http.ResponseWriter, r *http.Request, err error) {
	Error(w, r, err)
}

func UnprocessableEntity(w http.ResponseWriter, r *http.Request, err error) {
	Error(w, r, err)
}

func InternalServerError(w http.ResponseWriter, r *http.Request, err error) {
	Error(w, r, err)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if data != nil {
		_ = json.NewEncoder(w).Encode(data)
	}
}
