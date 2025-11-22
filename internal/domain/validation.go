package domain

import "fmt"

func (u User) Validate() error {
	if u.ID == "" {
		return NewValidationError("user_id", "must not be empty")
	}
	if u.Username == "" {
		return NewValidationError("username", "must not be empty")
	}
	if u.TeamName == "" {
		return NewValidationError("team_name", "must not be empty")
	}
	return nil
}

func (t Team) Validate() error {
	if t.Name == "" {
		return NewValidationError("team_name", "must not be empty")
	}

	for i, m := range t.Members {
		if err := m.Validate(); err != nil {
			return fmt.Errorf("member[%d]: %w", i, err)
		}
	}
	return nil
}

func (s PRStatus) Validate() error {
	switch s {
	case PRStatusOpen, PRStatusMerged:
		return nil
	default:
		return NewValidationError("status", "must be OPEN or MERGED")
	}
}

func (pr PullRequest) Validate() error {
	if pr.ID == "" {
		return NewValidationError("pull_request_id", "must not be empty")
	}
	if pr.Name == "" {
		return NewValidationError("pull_request_name", "must not be empty")
	}
	if pr.AuthorID == "" {
		return NewValidationError("author_id", "must not be empty")
	}
	if err := pr.Status.Validate(); err != nil {
		return err
	}
	if len(pr.AssignedReviewers) > 2 {
		return NewValidationError("assigned_reviewers", "must contain at most 2 reviewers")
	}

	seen := make(map[UserID]struct{}, len(pr.AssignedReviewers))
	for i, r := range pr.AssignedReviewers {
		if r == "" {
			return fmt.Errorf("assigned_reviewers[%d]: %w", i, NewValidationError("reviewer_id", "must not be empty"))
		}
		if _, ok := seen[r]; ok {
			return fmt.Errorf("assigned_reviewers[%d]: %w", i, NewValidationError("reviewer_id", "duplicate reviewer"))
		}
		seen[r] = struct{}{}
	}

	return nil
}
