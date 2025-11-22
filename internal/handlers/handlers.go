package handlers

import (
	"net/http"
	"pr-review-service/internal/models"
	"pr-review-service/internal/service"

	"github.com/gin-gonic/gin"
)

type ErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

type Handler struct {
	service *service.ReviewService
}

func NewHandler(service *service.ReviewService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) CreateTeam(c *gin.Context) {
	var team models.Team
	if err := c.BindJSON(&team); err != nil {
		c.JSON(http.StatusBadRequest, createError("INVALID_INPUT", err.Error()))
		return
	}

	if err := h.service.CreateTeam(team); err != nil {
		switch err {
		case service.ErrTeamExists:
			c.JSON(http.StatusConflict, createError("TEAM_EXISTS", err.Error()))
		default:
			c.JSON(http.StatusInternalServerError, createError("INTERNAL_ERROR", err.Error()))
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"team": team})
}

func (h *Handler) GetTeam(c *gin.Context) {
	teamName := c.Query("team_name")
	if teamName == "" {
		c.JSON(http.StatusBadRequest, createError("INVALID_INPUT", "team_name is required"))
		return
	}

	team, err := h.service.GetTeam(teamName)
	if err != nil {
		c.JSON(http.StatusNotFound, createError("NOT_FOUND", err.Error()))
		return
	}

	c.JSON(http.StatusOK, team)
}

func (h *Handler) SetUserActive(c *gin.Context) {
	var req struct {
		UserID   string `json:"user_id"`
		IsActive bool   `json:"is_active"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, createError("INVALID_INPUT", err.Error()))
		return
	}

	user, err := h.service.SetUserActive(req.UserID, req.IsActive)
	if err != nil {
		c.JSON(http.StatusNotFound, createError("NOT_FOUND", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

func (h *Handler) CreatePR(c *gin.Context) {
	var req struct {
		PullRequestID   string `json:"pull_request_id"`
		PullRequestName string `json:"pull_request_name"`
		AuthorID        string `json:"author_id"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, createError("INVALID_INPUT", err.Error()))
		return
	}

	pr, err := h.service.CreatePR(req.PullRequestID, req.PullRequestName, req.AuthorID)
	if err != nil {
		switch err {
		case service.ErrPRExists:
			c.JSON(http.StatusConflict, createError("PR_EXISTS", err.Error()))
		case service.ErrNotFound:
			c.JSON(http.StatusNotFound, createError("NOT_FOUND", err.Error()))
		default:
			c.JSON(http.StatusInternalServerError, createError("INTERNAL_ERROR", err.Error()))
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"pr": pr})
}

func (h *Handler) MergePR(c *gin.Context) {
	var req struct {
		PullRequestID string `json:"pull_request_id"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, createError("INVALID_INPUT", err.Error()))
		return
	}

	pr, err := h.service.MergePR(req.PullRequestID)
	if err != nil {
		c.JSON(http.StatusNotFound, createError("NOT_FOUND", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{"pr": pr})
}

func (h *Handler) ReassignReviewer(c *gin.Context) {
	var req struct {
		PullRequestID string `json:"pull_request_id"`
		OldUserID     string `json:"old_user_id"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, createError("INVALID_INPUT", err.Error()))
		return
	}

	pr, newUserID, err := h.service.ReassignReviewer(req.PullRequestID, req.OldUserID)
	if err != nil {
		switch err {
		case service.ErrPRMerged:
			c.JSON(http.StatusConflict, createError("PR_MERGED", err.Error()))
		case service.ErrNotAssigned:
			c.JSON(http.StatusConflict, createError("NOT_ASSIGNED", err.Error()))
		case service.ErrNoCandidate:
			c.JSON(http.StatusConflict, createError("NO_CANDIDATE", err.Error()))
		case service.ErrNotFound:
			c.JSON(http.StatusNotFound, createError("NOT_FOUND", err.Error()))
		default:
			c.JSON(http.StatusInternalServerError, createError("INTERNAL_ERROR", err.Error()))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"pr":          pr,
		"replaced_by": newUserID,
	})
}

func (h *Handler) GetUserReviews(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, createError("INVALID_INPUT", "user_id is required"))
		return
	}

	prs, err := h.service.GetUserReviews(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, createError("NOT_FOUND", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":       userID,
		"pull_requests": prs,
	})
}

func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "healthy"})
}

func createError(code, message string) ErrorResponse {
	var resp ErrorResponse
	resp.Error.Code = code
	resp.Error.Message = message
	return resp
}
