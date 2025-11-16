package v1

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/MatTwix/Pull-Request-Assigner/internal/metrics"
)

const (
	// Repository errors codes
	CodeTeamExists  = "TEAM_EXISTS"
	CodePRExists    = "PR_EXISTS"
	CodePRMerged    = "PR_MERGED"
	CodeNotAssigned = "NOT_ASSIGNED"
	CodeNoCandidate = "NO_CANDIDATE"
	CodeNotFound    = "NOT_FOUND"

	// Additional used error types codes
	CodeBadRequest          = "BAD_REQUEST"
	CodeInternalServerError = "INTERNAL_SERVER_ERROR"

	// Authorization errors codes
	CodeAdminAuthError = "ADMIN_AUTH"
	CodeUserAuthError  = "USER_AUTH"
)

var (
	ErrAdminAuthError = errors.New("incorrect admin key")
	ErrUserAuthError  = errors.New("incorrect user key")
)

type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func newErrorResponse(w http.ResponseWriter, status int, code, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	resp := ErrorResponse{
		Error: ErrorBody{
			Code:    code,
			Message: msg,
		},
	}

	if code != CodeInternalServerError {
		metrics.BusinessErrors.WithLabelValues(code).Inc()
	}

	_ = json.NewEncoder(w).Encode(resp)
}
