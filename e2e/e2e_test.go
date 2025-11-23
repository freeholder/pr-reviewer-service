package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"
)

type teamMemberDTO struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}

type teamAddRequest struct {
	TeamName string          `json:"team_name"`
	Members  []teamMemberDTO `json:"members"`
}

type pullRequestDTO struct {
	PullRequestID     string   `json:"pull_request_id"`
	PullRequestName   string   `json:"pull_request_name"`
	AuthorID          string   `json:"author_id"`
	Status            string   `json:"status"`
	AssignedReviewers []string `json:"assigned_reviewers"`
}

type createPRResponse struct {
	PR pullRequestDTO `json:"pr"`
}

type reassignRequest struct {
	PullRequestID string `json:"pull_request_id"`
	OldUserID     string `json:"old_user_id"`
}

type reassignResponse struct {
	PR         pullRequestDTO `json:"pr"`
	ReplacedBy string         `json:"replaced_by"`
}

type mergeRequest struct {
	PullRequestID string `json:"pull_request_id"`
}

type pullRequestShortDTO struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
	Status          string `json:"status"`
}

type getReviewResponse struct {
	UserID       string                `json:"user_id"`
	PullRequests []pullRequestShortDTO `json:"pull_requests"`
}

func baseURL() string {
	if v := os.Getenv("APP_BASE_URL"); v != "" {
		return v
	}
	return "http://localhost:8080"
}

func get(t *testing.T, path string) *http.Response {
	t.Helper()

	url := baseURL() + path
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
	return resp
}

func postJSON(t *testing.T, path string, body any) *http.Response {
	t.Helper()

	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal body for %s: %v", path, err)
	}

	req, err := http.NewRequest(http.MethodPost, baseURL()+path, bytes.NewReader(data))
	if err != nil {
		t.Fatalf("new request %s: %v", path, err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST %s: %v", path, err)
	}
	return resp
}

func decodeJSON(t *testing.T, r io.Reader, dst any) {
	t.Helper()
	if err := json.NewDecoder(r).Decode(dst); err != nil {
		t.Fatalf("decode json: %v", err)
	}
}

func TestEndToEnd_Create_Reassign_Merge(t *testing.T) {
	resp := get(t, "/health")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Skipf("service is not healthy (status %d): %s", resp.StatusCode, string(body))
	}

	now := time.Now().UnixNano()
	teamName := fmt.Sprintf("e2e-team-%d", now)
	u1 := fmt.Sprintf("e2e-u1-%d", now)
	u2 := fmt.Sprintf("e2e-u2-%d", now)
	u3 := fmt.Sprintf("e2e-u3-%d", now)
	u4 := fmt.Sprintf("e2e-u4-%d", now)
	prID := fmt.Sprintf("e2e-pr-%d", now)

	t.Run("create team", func(t *testing.T) {
		req := teamAddRequest{
			TeamName: teamName,
			Members: []teamMemberDTO{
				{UserID: u1, Username: "Alice", IsActive: true},
				{UserID: u2, Username: "Bob", IsActive: true},
				{UserID: u3, Username: "Charlie", IsActive: true},
				{UserID: u4, Username: "Diana", IsActive: true},
			},
		}

		resp := postJSON(t, "/team/add", req)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("expected 201 from /team/add, got %d: %s", resp.StatusCode, string(body))
		}
	})

	var createdPR createPRResponse

	t.Run("create pr", func(t *testing.T) {
		body := map[string]any{
			"pull_request_id":   prID,
			"pull_request_name": "E2E test PR",
			"author_id":         u1,
		}

		resp := postJSON(t, "/pullRequest/create", body)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			data, _ := io.ReadAll(resp.Body)
			t.Fatalf("expected 201 from /pullRequest/create, got %d: %s", resp.StatusCode, string(data))
		}

		decodeJSON(t, resp.Body, &createdPR)

		if createdPR.PR.Status != "OPEN" {
			t.Fatalf("expected PR status OPEN, got %s", createdPR.PR.Status)
		}
		if createdPR.PR.AuthorID != u1 {
			t.Fatalf("expected author_id %s, got %s", u1, createdPR.PR.AuthorID)
		}
		if len(createdPR.PR.AssignedReviewers) == 0 {
			t.Fatalf("expected at least one assigned reviewer")
		}
	})

	var oldReviewer, newReviewer string
	t.Run("reassign reviewer", func(t *testing.T) {
		oldReviewer = createdPR.PR.AssignedReviewers[0]

		req := reassignRequest{
			PullRequestID: prID,
			OldUserID:     oldReviewer,
		}

		resp := postJSON(t, "/pullRequest/reassign", req)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			data, _ := io.ReadAll(resp.Body)
			t.Fatalf("expected 200 from /pullRequest/reassign, got %d: %s", resp.StatusCode, string(data))
		}

		var rr reassignResponse
		decodeJSON(t, resp.Body, &rr)

		if rr.ReplacedBy == "" {
			t.Fatalf("expected non-empty replaced_by")
		}
		if rr.ReplacedBy == oldReviewer {
			t.Fatalf("replaced_by must differ from old reviewer")
		}

		newReviewer = rr.ReplacedBy
	})

	t.Run("merge pr", func(t *testing.T) {
		req := mergeRequest{PullRequestID: prID}

		resp := postJSON(t, "/pullRequest/merge", req)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			data, _ := io.ReadAll(resp.Body)
			t.Fatalf("expected 200 from /pullRequest/merge, got %d: %s", resp.StatusCode, string(data))
		}

		var mr struct {
			PR pullRequestDTO `json:"pr"`
		}
		decodeJSON(t, resp.Body, &mr)

		if mr.PR.Status != "MERGED" {
			t.Fatalf("expected PR status MERGED, got %s", mr.PR.Status)
		}
	})

	t.Run("check reviewers list", func(t *testing.T) {
		resp := get(t, "/users/getReview?user_id="+newReviewer)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			data, _ := io.ReadAll(resp.Body)
			t.Fatalf("expected 200 from /users/getReview (newReviewer), got %d: %s", resp.StatusCode, string(data))
		}

		var grNew getReviewResponse
		decodeJSON(t, resp.Body, &grNew)

		found := false
		for _, pr := range grNew.PullRequests {
			if pr.PullRequestID == prID {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected PR %s in getReview for new reviewer %s", prID, newReviewer)
		}

		resp2 := get(t, "/users/getReview?user_id="+oldReviewer)
		defer resp2.Body.Close()

		if resp2.StatusCode != http.StatusOK {
			data, _ := io.ReadAll(resp2.Body)
			t.Fatalf("expected 200 from /users/getReview (oldReviewer), got %d: %s", resp2.StatusCode, string(data))
		}

		var grOld getReviewResponse
		decodeJSON(t, resp2.Body, &grOld)

		for _, pr := range grOld.PullRequests {
			if pr.PullRequestID == prID {
				t.Fatalf("did not expect PR %s in getReview for old reviewer %s", prID, oldReviewer)
			}
		}
	})
}
