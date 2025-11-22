package domain

import "time"

type (
	UserID        string
	TeamName      string
	PullRequestID string
)

type User struct {
	ID       UserID
	Username string
	TeamName TeamName
	IsActive bool
}

type Team struct {
	Name    TeamName
	Members []User
}

type PRStatus string

const (
	PRStatusOpen   PRStatus = "OPEN"
	PRStatusMerged PRStatus = "MERGED"
)

type PullRequest struct {
	ID                PullRequestID
	Name              string
	AuthorID          UserID
	Status            PRStatus
	AssignedReviewers []UserID
	CreatedAt         *time.Time
	MergedAt          *time.Time
}
