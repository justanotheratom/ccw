package github

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os/exec"
	"regexp"
	"strings"
)

var githubURLPatterns = []*regexp.Regexp{
	regexp.MustCompile(`^https://github\.com/`),
	regexp.MustCompile(`^git@github\.com:`),
	regexp.MustCompile(`^ssh://git@github\.com/`),
}

// Client provides GitHub integration via the gh CLI.
type Client struct {
	repoPath string

	// Cached state
	isGitHubRepo *bool
}

// NewClient creates a new GitHub client for the given repo path.
func NewClient(repoPath string) *Client {
	return &Client{repoPath: repoPath}
}

// IsGitHubRepo checks if the repo's origin remote points to GitHub.
func (c *Client) IsGitHubRepo() bool {
	if c.isGitHubRepo != nil {
		return *c.isGitHubRepo
	}

	cmd := exec.Command("git", "-C", c.repoPath, "remote", "get-url", "origin")
	out, err := cmd.Output()
	if err != nil {
		result := false
		c.isGitHubRepo = &result
		return false
	}

	url := strings.TrimSpace(string(out))
	for _, pattern := range githubURLPatterns {
		if pattern.MatchString(url) {
			result := true
			c.isGitHubRepo = &result
			return true
		}
	}

	result := false
	c.isGitHubRepo = &result
	return false
}

// ErrNotAuthenticated is returned when gh CLI is not authenticated.
var ErrNotAuthenticated = errors.New("gh CLI is not authenticated; run `gh auth login`")

// CheckAuthenticated verifies that gh CLI is authenticated for github.com.
func CheckAuthenticated() error {
	cmd := exec.Command("gh", "auth", "status", "--hostname", "github.com")
	if err := cmd.Run(); err != nil {
		return ErrNotAuthenticated
	}
	return nil
}

type prViewResult struct {
	State    string `json:"state"`
	MergedAt string `json:"mergedAt"`
}

// IsPRMerged checks if a branch was merged via PR.
// Returns (merged, found, error) where:
// - found=false means no PR exists for this branch (not an error)
// - found=true means a PR was found, and merged indicates its state
func (c *Client) IsPRMerged(ctx context.Context, branch string) (merged bool, found bool, err error) {
	cmd := exec.CommandContext(ctx, "gh", "pr", "view", branch,
		"--json", "state,mergedAt")
	cmd.Dir = c.repoPath

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// Check if it's "no pull requests found" (exit code 1)
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			stderrStr := stderr.String()
			if strings.Contains(stderrStr, "no pull requests found") ||
				strings.Contains(stderrStr, "Could not resolve") {
				return false, false, nil // No PR for this branch
			}
		}
		return false, false, err
	}

	var result prViewResult
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		return false, false, err
	}

	return result.State == "MERGED", true, nil
}
