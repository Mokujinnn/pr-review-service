package storage

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"pr-review-service/internal/models"

	_ "github.com/lib/pq"
)

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage() (*PostgresStorage, error) {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "password")
	dbname := getEnv("DB_NAME", "pr_review_service")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	for i := range 30 {
		err = db.Ping()
		if err == nil {
			break
		}
		log.Printf("Waiting for database... (attempt %d/30)", i+1)
		time.Sleep(1 * time.Second)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database after 30 attempts: %v", err)
	}

	log.Println("Successfully connected to database")
	return &PostgresStorage{db: db}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func (s *PostgresStorage) CreateTeam(team models.Team) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec("INSERT INTO teams (team_name) VALUES ($1) ON CONFLICT (team_name) DO NOTHING", team.TeamName)
	if err != nil {
		return err
	}

	for _, member := range team.Members {
		_, err = tx.Exec(`
			INSERT INTO users (user_id, username, team_name, is_active) 
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (user_id) 
			DO UPDATE SET username = $2, team_name = $3, is_active = $4, updated_at = CURRENT_TIMESTAMP`,
			member.UserID, member.Username, team.TeamName, member.IsActive)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *PostgresStorage) GetTeam(teamName string) (*models.Team, error) {
	var team models.Team
	team.TeamName = teamName

	rows, err := s.db.Query(`
		SELECT user_id, username, is_active 
		FROM users 
		WHERE team_name = $1 
		ORDER BY user_id`, teamName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var member models.User
		if err := rows.Scan(&member.UserID, &member.Username, &member.IsActive); err != nil {
			return nil, err
		}
		member.TeamName = teamName
		team.Members = append(team.Members, member)
	}

	if len(team.Members) == 0 {
		return nil, errors.New("team not found")
	}

	return &team, nil
}

func (s *PostgresStorage) GetUser(userID string) (*models.User, error) {
	var user models.User
	err := s.db.QueryRow(`
		SELECT user_id, username, team_name, is_active 
		FROM users 
		WHERE user_id = $1`, userID).Scan(&user.UserID, &user.Username, &user.TeamName, &user.IsActive)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

func (s *PostgresStorage) UpdateUser(user *models.User) error {
	_, err := s.db.Exec(`
		UPDATE users 
		SET username = $1, team_name = $2, is_active = $3, updated_at = CURRENT_TIMESTAMP 
		WHERE user_id = $4`,
		user.Username, user.TeamName, user.IsActive, user.UserID)
	return err
}

func (s *PostgresStorage) GetActiveTeamMembers(teamName string) ([]models.User, error) {
	rows, err := s.db.Query(`
		SELECT user_id, username, team_name, is_active 
		FROM users 
		WHERE team_name = $1 AND is_active = true 
		ORDER BY user_id`, teamName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []models.User
	for rows.Next() {
		var member models.User
		if err := rows.Scan(&member.UserID, &member.Username, &member.TeamName, &member.IsActive); err != nil {
			return nil, err
		}
		members = append(members, member)
	}

	return members, nil
}

func (s *PostgresStorage) CreatePR(pr *models.PullRequest) error {
	reviewersJSON, err := json.Marshal(pr.AssignedReviewers)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(`
		INSERT INTO pull_requests 
		(pull_request_id, pull_request_name, author_id, status, assigned_reviewers, created_at, merged_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		pr.PullRequestID, pr.PullRequestName, pr.AuthorID, pr.Status, reviewersJSON, pr.CreatedAt, pr.MergedAt)
	return err
}

func (s *PostgresStorage) GetPR(prID string) (*models.PullRequest, error) {
	var pr models.PullRequest
	var reviewersJSON []byte
	var mergedAt sql.NullTime

	err := s.db.QueryRow(`
		SELECT pull_request_id, pull_request_name, author_id, status, assigned_reviewers, created_at, merged_at 
		FROM pull_requests 
		WHERE pull_request_id = $1`, prID).Scan(
		&pr.PullRequestID, &pr.PullRequestName, &pr.AuthorID, &pr.Status, &reviewersJSON, &pr.CreatedAt, &mergedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("PR not found")
		}
		return nil, err
	}

	if err := json.Unmarshal(reviewersJSON, &pr.AssignedReviewers); err != nil {
		return nil, err
	}

	if mergedAt.Valid {
		pr.MergedAt = &mergedAt.Time
	}

	return &pr, nil
}

func (s *PostgresStorage) UpdatePR(pr *models.PullRequest) error {
	reviewersJSON, err := json.Marshal(pr.AssignedReviewers)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(`
		UPDATE pull_requests 
		SET pull_request_name = $1, author_id = $2, status = $3, assigned_reviewers = $4, 
		    merged_at = $5, updated_at = CURRENT_TIMESTAMP 
		WHERE pull_request_id = $6`,
		pr.PullRequestName, pr.AuthorID, pr.Status, reviewersJSON, pr.MergedAt, pr.PullRequestID)
	return err
}

func (s *PostgresStorage) GetPRsByReviewer(userID string) ([]models.PullRequestShort, error) {
	rows, err := s.db.Query(`
		SELECT pull_request_id, pull_request_name, author_id, status 
		FROM pull_requests 
		WHERE assigned_reviewers @> $1 
		ORDER BY created_at DESC`, fmt.Sprintf(`["%s"]`, userID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prs []models.PullRequestShort
	for rows.Next() {
		var pr models.PullRequestShort
		if err := rows.Scan(&pr.PullRequestID, &pr.PullRequestName, &pr.AuthorID, &pr.Status); err != nil {
			return nil, err
		}
		prs = append(prs, pr)
	}

	return prs, nil
}

func (s *PostgresStorage) Close() error {
	return s.db.Close()
}
