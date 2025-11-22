package storage

import "pr-review-service/internal/models"

type Storage interface {
	CreateTeam(team models.Team) error
	GetTeam(teamName string) (*models.Team, error)

	GetUser(userID string) (*models.User, error)
	UpdateUser(user *models.User) error
	GetActiveTeamMembers(teamName string) ([]models.User, error)

	CreatePR(pr *models.PullRequest) error
	GetPR(prID string) (*models.PullRequest, error)
	UpdatePR(pr *models.PullRequest) error
	GetPRsByReviewer(userID string) ([]models.PullRequestShort, error)

	Close() error
}
