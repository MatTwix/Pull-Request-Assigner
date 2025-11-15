package v1

import (
	"encoding/json"
	"net/http"

	"github.com/MatTwix/Pull-Request-Assigner/internal/repo/repoerrs"
	"github.com/MatTwix/Pull-Request-Assigner/internal/service"
	"github.com/MatTwix/Pull-Request-Assigner/pkg/logger"
	"github.com/MatTwix/Pull-Request-Assigner/pkg/utils"
)

type pullRequestRoutes struct {
	prService service.PullRequest
	logger    logger.Logger
}

func newPullRequestRoutes(prService service.PullRequest, logger logger.Logger) *pullRequestRoutes {
	prr := &pullRequestRoutes{
		prService: prService,
		logger:    logger,
	}

	return prr
}

type createPRRequest struct {
	PullRequestID   string `json:"pull_request_id" validate:"required"`
	PullRequestName string `json:"pull_request_name" validate:"required"`
	AuthorID        string `json:"author_id" validate:"required"`
}

// @Summary Создать пулл реквест
// @Description Создает пулл реквест и автоматически назначает до 2 ревьюверов из команды автора
// @Tags PullRequests
// @Accept json
// @Produce json
// @Param request body createPRRequest true "PullRequest payload"
// @Success 201 {object} service.PullRequestCreateOutput
// @Failure 400 {object} ErrorResponse "Неверное тело запроса"
// @Failure 401 {object} ErrorResponse "Ошибка авторизации"
// @Failure 404 {object} ErrorResponse "Автор/команда не найдены"
// @Failure 409 {object} ErrorResponse "PR уже существует"
// @Failure 500 {object} ErrorResponse "Внутренняя ошибка сервера"
// @Security ApiKeyAuth
// @Router /pullRequest/create [post]
func (prr *pullRequestRoutes) create(w http.ResponseWriter, r *http.Request) {
	var req createPRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		newErrorResponse(w, http.StatusBadRequest, CodeBadRequest, "invalid request body")
		return
	}

	if err := utils.ValidateStruct(&req); err != nil {
		newErrorResponse(w, http.StatusBadRequest, CodeBadRequest, "invalid request body")
		return
	}

	input := service.PullRequestCreateInput{
		PullRequestID:   req.PullRequestID,
		PullRequestName: req.PullRequestName,
		AuthorID:        req.AuthorID,
	}

	pullRequest, err := prr.prService.CreatePR(r.Context(), input)
	if err != nil {
		switch err {
		case repoerrs.ErrNotFound:
			newErrorResponse(w, http.StatusNotFound, CodeNotFound, err.Error())
			return
		case repoerrs.ErrAlreadyExists:
			newErrorResponse(w, http.StatusConflict, CodePRExists, "PR is already exists")
			return
		default:
			newErrorResponse(w, http.StatusInternalServerError, CodeInternalServerError, "failed to create pull request")
			prr.logger.Error("failed to create pull request", map[string]any{
				"pr_id":     req.PullRequestID,
				"pr_name":   req.PullRequestName,
				"author_id": req.AuthorID,
				"error":     err,
			})
			return
		}
	}

	newSuccessResponse(w, http.StatusCreated, pullRequest)
}

type mergePRRequest struct {
	PullRequestID string `json:"pull_request_id" validate:"required"`
}

// @Summary Установить статус пулл реквеста "MERGED"
// @Description Идемпотентно устанавливает статус пулл реквеста "MERGED"
// @Tags PullRequests
// @Accept json
// @Produce json
// @Param request body mergePRRequest true "Merge payload"
// @Success 200 {object} service.PullRequestMergeOutput
// @Failure 400 {object} ErrorResponse "Неверное тело запроса"
// @Failure 401 {object} ErrorResponse "Ошибка авторизации"
// @Failure 404 {object} ErrorResponse "PR не найден"
// @Failure 500 {object} ErrorResponse "Внутренняя ошибка сервера"
// @Security ApiKeyAuth
// @Router /pullRequest/merge [post]
func (prr *pullRequestRoutes) merge(w http.ResponseWriter, r *http.Request) {
	var req mergePRRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		newErrorResponse(w, http.StatusBadRequest, CodeBadRequest, "invalid request body")
		return
	}

	if err := utils.ValidateStruct(req); err != nil {
		newErrorResponse(w, http.StatusBadRequest, CodeBadRequest, "invalid request body")
		return
	}

	pullRequest, err := prr.prService.MergePR(r.Context(), req.PullRequestID)
	if err != nil {
		switch err {
		case repoerrs.ErrNotFound:
			newErrorResponse(w, http.StatusNotFound, CodeNotFound, err.Error())
			return
		default:
			newErrorResponse(w, http.StatusInternalServerError, CodeInternalServerError, "failed to merge pull request")
			prr.logger.Error("failed to merge pull request", map[string]any{
				"pr_id": req.PullRequestID,
				"error": err,
			})
			return
		}
	}

	newSuccessResponse(w, http.StatusOK, pullRequest)
}

type reassignRequest struct {
	PullRequestID string `json:"pull_request_id"`
	OldUserID     string `json:"old_user_id"`
}

// @Summary Переназначить ревьювера
// @Description Меняет одного ревьювера на рандомного из его команды
// @Tags PullRequests
// @Accept json
// @Produce json
// @Param request body reassignRequest true "Reassign payload"
// @Success 200 {object} service.PullRequestReassignOutput
// @Failure 400 {object} ErrorResponse "Неверное тело запроса"
// @Failure 401 {object} ErrorResponse "Ошибка авторизации"
// @Failure 404 {object} ErrorResponse "PR или пользователь не найден"
// @Failure 409 {object} ErrorResponse "Нарушение правил переназначения"
// @Failure 500 {object} ErrorResponse "Внутренняя ошибка сервера"
// @Security ApiKeyAuth
// @Router /pullRequest/reassign [post]
func (prr *pullRequestRoutes) reassign(w http.ResponseWriter, r *http.Request) {
	var req reassignRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		newErrorResponse(w, http.StatusBadRequest, CodeBadRequest, "invalid request body")
		return
	}

	if err := utils.ValidateStruct(req); err != nil {
		newErrorResponse(w, http.StatusBadRequest, CodeBadRequest, "invalid request body")
		return
	}

	response, err := prr.prService.ReassignReviewer(r.Context(), req.PullRequestID, req.OldUserID)
	if err != nil {
		switch err {
		case repoerrs.ErrNotFound:
			newErrorResponse(w, http.StatusNotFound, CodeNotFound, "pull request not found")
			return
		case repoerrs.ErrUserNotFound:
			newErrorResponse(w, http.StatusNotFound, CodeNotFound, err.Error())
			return
		case repoerrs.ErrReassignAfterMerge:
			newErrorResponse(w, http.StatusConflict, CodePRMerged, err.Error())
			return
		case repoerrs.ErrNotAssigned:
			newErrorResponse(w, http.StatusConflict, CodeNotAssigned, err.Error())
			return
		case repoerrs.ErrNoCandidate:
			newErrorResponse(w, http.StatusConflict, CodeNoCandidate, err.Error())
			return
		default:
			newErrorResponse(w, http.StatusInternalServerError, CodeInternalServerError, "failed to reassign reviewer")
			prr.logger.Error("failed to reassign reviewer", map[string]any{
				"pr_id":       req.PullRequestID,
				"old_user_id": req.OldUserID,
				"error":       err,
			})
			return
		}
	}

	newSuccessResponse(w, http.StatusOK, response)
}
