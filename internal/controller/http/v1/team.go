package v1

import (
	"encoding/json"
	"net/http"

	"github.com/MatTwix/Pull-Request-Assigner/internal/repo/repoerrs"
	"github.com/MatTwix/Pull-Request-Assigner/internal/service"
	"github.com/MatTwix/Pull-Request-Assigner/pkg/logger"
	"github.com/MatTwix/Pull-Request-Assigner/pkg/utils"
)

type teamRoutes struct {
	teamService service.Team
	logger      logger.Logger
}

func newTeamRoutes(teamService service.Team, logger logger.Logger) *teamRoutes {
	tr := &teamRoutes{
		teamService: teamService,
		logger:      logger,
	}

	return tr
}

type addTeamRequest struct {
	TeamName string       `json:"team_name" validate:"required"`
	Members  []teamMember `json:"members" validate:"required,dive"`
}

type teamMember struct {
	UserID   string `json:"user_id" validate:"required"`
	Username string `json:"username" validate:"required"`
	IsActive bool   `json:"is_active" validate:"required"`
}

// @Summary Создать команду с участниками
// @Description Создает или обновляет команду с указанными пользователями
// @Tags Teams
// @Accept json
// @Produce json
// @Param request body addTeamRequest true "Team payload"
// @Success 201 {object} service.TeamAddOutput
// @Failure 400 {object} ErrorResponse "Команда уже существует или неверное тело запроса"
// @Failure 401 {object} ErrorResponse "Ошибка авторизации"
// @Failure 500 {object} ErrorResponse "Внутренняя ошибка сервера"
// @Security ApiKeyAuth
// @Router /team/add [post]
func (tr *teamRoutes) add(w http.ResponseWriter, r *http.Request) {
	var req addTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		newErrorResponse(w, http.StatusBadRequest, CodeBadRequest, "invalid request body")
		return
	}

	if err := utils.ValidateStruct(req); err != nil {
		newErrorResponse(w, http.StatusBadRequest, CodeBadRequest, "invalid request body")
		return
	}

	input := service.TeamAddInput{TeamName: req.TeamName}

	for _, member := range req.Members {
		input.Members = append(input.Members, service.TeamInputMember{
			UserID:   member.UserID,
			Username: member.Username,
			IsActive: member.IsActive,
		})
	}

	team, err := tr.teamService.AddTeam(r.Context(), input)
	if err != nil {
		switch err {
		case repoerrs.ErrAlreadyExists:
			newErrorResponse(w, http.StatusBadRequest, CodeTeamExists, "team_name already exists")
			return
		default:
			newErrorResponse(w, http.StatusInternalServerError, CodeInternalServerError, "failed to create team")
			tr.logger.Error("failed to create team", map[string]any{
				"team_name":      req.TeamName,
				"members_amount": len(req.Members),
				"error":          err,
			})
			return
		}
	}

	newSuccessResponse(w, http.StatusCreated, team)
}

// @Summary Получить команду с участниками
// @Description Возвращает информацию о команде и ее пользователях
// @Tags Teams
// @Accept json
// @Produce json
// @Param team_name query string true "Имя команды"
// @Success 200 {object} service.TeamGetOutput
// @Failure 400 {object} ErrorResponse "Неверное имя команды"
// @Failure 401 {object} ErrorResponse "Ошибка авторизации"
// @Failure 404 {object} ErrorResponse "Команда не найдена"
// @Failure 500 {object} ErrorResponse "Внутренняя ошибка сервера"
// @Router /team/get [get]
func (tr *teamRoutes) get(w http.ResponseWriter, r *http.Request) {
	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		newErrorResponse(w, http.StatusBadRequest, CodeBadRequest, "invalid team_name")
		return
	}

	team, err := tr.teamService.GetTeamByName(r.Context(), teamName)
	if err != nil {
		switch err {
		case repoerrs.ErrNotFound:
			newErrorResponse(w, http.StatusNotFound, CodeNotFound, err.Error())
			return
		default:
			newErrorResponse(w, http.StatusInternalServerError, CodeInternalServerError, "failed to get team by name")
			tr.logger.Error("failed to get team by name", map[string]any{
				"team_name": teamName,
				"error":     err,
			})
			return
		}
	}

	newSuccessResponse(w, http.StatusOK, team)
}

type deactivateTeamRequest struct {
	TeamName string `json:"team_name" validate:"required"`
}

// @Summary Деактивация всех членов команды
// @Description Быстрый метод для массовой деактивации членов определенной команды
// @Tags Teams
// @Accept json
// @Produce json
// @Param request body deactivateTeamRequest true "Team to deactivate name"
// @Success 200 {object} service.TeamSetIsActiveTeamOutput
// @Failure 400 {object} ErrorResponse "Неверное тело запроса"
// @Failure 404 {object} ErrorResponse "Команда не найдена"
// @Failure 500 {object} ErrorResponse "Внутренняя ошибка сервера"
// @Router /teams/deactivate [post]
func (tr *teamRoutes) deactivateTeam(w http.ResponseWriter, r *http.Request) {
	var req deactivateTeamRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		newErrorResponse(w, http.StatusBadRequest, CodeBadRequest, "invalid request body")
		return
	}

	if err := utils.ValidateStruct(req); err != nil || req.TeamName == "" {
		newErrorResponse(w, http.StatusBadRequest, CodeBadRequest, "invalid request body")
		return
	}

	usersDeactivated, err := tr.teamService.SetIsActiveTeam(r.Context(), req.TeamName, false)
	if err != nil {
		switch err {
		case repoerrs.ErrNotFound:
			newErrorResponse(w, http.StatusNotFound, CodeNotFound, err.Error())
			return
		default:
			newErrorResponse(w, http.StatusInternalServerError, CodeInternalServerError, "failed to deactivate users")
			tr.logger.Error("failed to deactivate users", map[string]any{
				"team_name": len(req.TeamName),
				"error":     err,
			})
			return
		}
	}

	newSuccessResponse(w, http.StatusOK, usersDeactivated)
}
