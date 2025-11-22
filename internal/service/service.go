package service

import (
	"errors"
	"math/rand"
	"pr-review-service/internal/models"
	"pr-review-service/internal/storage"
	"time"
)

var (
	ErrTeamExists  = errors.New("team already exists")
	ErrPRExists    = errors.New("PR already exists")
	ErrPRMerged    = errors.New("PR is merged")
	ErrNotAssigned = errors.New("reviewer not assigned")
	ErrNoCandidate = errors.New("no active replacement candidate")
	ErrNotFound    = errors.New("resource not found")
)

type ReviewService struct {
	store storage.Storage
}

func NewReviewService(store storage.Storage) *ReviewService {
	return &ReviewService{store: store}
}

func (s *ReviewService) CreateTeam(team models.Team) error {
	existing, _ := s.store.GetTeam(team.TeamName)
	if existing != nil {
		return ErrTeamExists
	}

	return s.store.CreateTeam(team)
}

func (s *ReviewService) GetTeam(teamName string) (*models.Team, error) {
	team, err := s.store.GetTeam(teamName)
	if err != nil {
		return nil, ErrNotFound
	}
	return team, nil
}

func (s *ReviewService) SetUserActive(userID string, isActive bool) (*models.User, error) {
	user, err := s.store.GetUser(userID)
	if err != nil {
		return nil, ErrNotFound
	}

	user.IsActive = isActive
	if err := s.store.UpdateUser(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *ReviewService) CreatePR(prID, prName, authorID string) (*models.PullRequest, error) {
	existing, _ := s.store.GetPR(prID)
	if existing != nil {
		return nil, ErrPRExists
	}

	author, err := s.store.GetUser(authorID)
	if err != nil {
		return nil, ErrNotFound
	}

	teamMembers, err := s.store.GetActiveTeamMembers(author.TeamName)
	if err != nil {
		return nil, err
	}

	var candidates []models.User
	for _, member := range teamMembers {
		if member.UserID != authorID {
			candidates = append(candidates, member)
		}
	}

	reviewers := s.selectReviewers(candidates, 2)

	reviewerIDs := make([]string, len(reviewers))
	for i, reviewer := range reviewers {
		reviewerIDs[i] = reviewer.UserID
	}

	pr := &models.PullRequest{
		PullRequestID:     prID,
		PullRequestName:   prName,
		AuthorID:          authorID,
		Status:            "OPEN",
		AssignedReviewers: reviewerIDs,
		CreatedAt:         time.Now(),
	}

	if err := s.store.CreatePR(pr); err != nil {
		return nil, err
	}

	return pr, nil
}

func (s *ReviewService) MergePR(prID string) (*models.PullRequest, error) {
	pr, err := s.store.GetPR(prID)
	if err != nil {
		return nil, ErrNotFound
	}

	if pr.Status == "MERGED" {
		return pr, nil
	}

	now := time.Now()
	pr.Status = "MERGED"
	pr.MergedAt = &now

	if err := s.store.UpdatePR(pr); err != nil {
		return nil, err
	}

	return pr, nil
}

func (s *ReviewService) ReassignReviewer(prID, oldUserID string) (*models.PullRequest, string, error) {
	pr, err := s.store.GetPR(prID)
	if err != nil {
		return nil, "", ErrNotFound
	}

	if pr.Status == "MERGED" {
		return nil, "", ErrPRMerged
	}

	found := false
	for _, reviewer := range pr.AssignedReviewers {
		if reviewer == oldUserID {
			found = true
			break
		}
	}
	if !found {
		return nil, "", ErrNotAssigned
	}

	oldReviewer, err := s.store.GetUser(oldUserID)
	if err != nil {
		return nil, "", ErrNotFound
	}

	teamMembers, err := s.store.GetActiveTeamMembers(oldReviewer.TeamName)
	if err != nil {
		return nil, "", err
	}

	var candidates []models.User
	for _, member := range teamMembers {
		if member.UserID != oldUserID && member.UserID != pr.AuthorID {
			candidates = append(candidates, member)
		}
	}

	if len(candidates) == 0 {
		return nil, "", ErrNoCandidate
	}

	newReviewer := candidates[rand.Intn(len(candidates))]

	for i, reviewer := range pr.AssignedReviewers {
		if reviewer == oldUserID {
			pr.AssignedReviewers[i] = newReviewer.UserID
			break
		}
	}

	if err := s.store.UpdatePR(pr); err != nil {
		return nil, "", err
	}

	return pr, newReviewer.UserID, nil
}

func (s *ReviewService) GetUserReviews(userID string) ([]models.PullRequestShort, error) {
	_, err := s.store.GetUser(userID)
	if err != nil {
		return nil, ErrNotFound
	}

	return s.store.GetPRsByReviewer(userID)
}

func (s *ReviewService) selectReviewers(candidates []models.User, max int) []models.User {
	if len(candidates) == 0 {
		return []models.User{}
	}

	if len(candidates) <= max {
		return candidates
	}

	shuffled := make([]models.User, len(candidates))
	copy(shuffled, candidates)
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	return shuffled[:max]
}
