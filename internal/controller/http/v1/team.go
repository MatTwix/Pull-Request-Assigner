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
