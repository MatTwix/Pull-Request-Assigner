package v1

import (
	"encoding/json"
	"net/http"

	"github.com/MatTwix/Pull-Request-Assigner/internal/repo/repoerrs"
	"github.com/MatTwix/Pull-Request-Assigner/internal/service"
	"github.com/MatTwix/Pull-Request-Assigner/pkg/logger"
	"github.com/MatTwix/Pull-Request-Assigner/pkg/utils"
)

type userRoutes struct {
	userService service.User
	logger      logger.Logger
}

func newUserRoutes(userService service.User, logger logger.Logger) *userRoutes {
	ur := &userRoutes{
		userService: userService,
		logger:      logger,
	}

	return ur
}

type setIsActiveRequest struct {
	UserID   string `json:"user_id" validate:"required"`
	IsActive bool   `json:"is_active"`
}

func (ur *userRoutes) setIsActive(w http.ResponseWriter, r *http.Request) {
	var req setIsActiveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		newErrorResponse(w, http.StatusBadRequest, CodeBadRequest, "invalid request body")
		return
	}

	if err := utils.ValidateStruct(req); err != nil {
		newErrorResponse(w, http.StatusBadRequest, CodeBadRequest, "invalid request body")
		return
	}

	user, err := ur.userService.SetIsActive(r.Context(), req.UserID, req.IsActive)
	if err != nil {
		switch err {
		case repoerrs.ErrNotFound:
			newErrorResponse(w, http.StatusNotFound, CodeNotFound, err.Error())
			return
		default:
			newErrorResponse(w, http.StatusInternalServerError, CodeInternalServerError, "failed to set is_active prop")
			ur.logger.Error("failed to set is_active prop", map[string]any{
				"user_id":   req.UserID,
				"is_active": req.IsActive,
				"error":     err,
			})
			return
		}
	}

	newSuccessResponse(w, http.StatusOK, user)
}

func (ur *userRoutes) getReview(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		newErrorResponse(w, http.StatusBadRequest, CodeBadRequest, "invalid user_id")
		return
	}

	pullRequests, err := ur.userService.GetReview(r.Context(), userID)
	if err != nil {
		newErrorResponse(w, http.StatusInternalServerError, CodeInternalServerError, "failed to get reviewed pull requests")
		ur.logger.Error("failed to get reviewed pull requests", map[string]any{
			"user_id": userID,
			"error":   err,
		})
		return
	}

	newSuccessResponse(w, http.StatusOK, pullRequests)
}
