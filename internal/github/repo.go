package github

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

	gogithub "github.com/google/go-github/v57/github"
	"golang.org/x/oauth2"
)

// Client wraps the GitHub API client.
type Client struct {
	gh    *gogithub.Client
	owner string
}

// NewClient creates a GitHub client from a personal access token.
// Token is read from GITHUB_TOKEN env var if token is empty.
func NewClient(token string) (*Client, error) {
	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
	}
	if token == "" {
		return nil, fmt.Errorf("GitHub token required: set GITHUB_TOKEN env var or pass --token flag")
	}

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(context.Background(), ts)
	gh := gogithub.NewClient(tc)

	// Verify token and get authenticated user
	user, _, err := gh.Users.Get(context.Background(), "")
	if err != nil {
		return nil, fmt.Errorf("GitHub auth failed: %w", err)
	}

	return &Client{
		gh:    gh,
		owner: user.GetLogin(),
	}, nil
}

// Owner returns the authenticated GitHub username.
func (c *Client) Owner() string {
	return c.owner
}

// CreateOrGetRepo creates a new GitHub repository, or returns the existing one.
func (c *Client) CreateOrGetRepo(ctx context.Context, name, description string, private bool) (*gogithub.Repository, error) {
	// Check if repo already exists
	repo, _, err := c.gh.Repositories.Get(ctx, c.owner, name)
	if err == nil {
		return repo, nil
	}

	// Create the repository
	repo, _, err = c.gh.Repositories.Create(ctx, "", &gogithub.Repository{
		Name:        gogithub.String(name),
		Description: gogithub.String(description),
		Private:     gogithub.Bool(private),
		AutoInit:    gogithub.Bool(true),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create repository: %w", err)
	}

	return repo, nil
}

// UploadFile creates or updates a file in the repository.
func (c *Client) UploadFile(ctx context.Context, repoName, filePath, localPath, commitMsg string) error {
	content, err := os.ReadFile(localPath)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", localPath, err)
	}

	// Check if file already exists (to get its SHA for updates)
	var fileSHA *string
	existing, _, _, err := c.gh.Repositories.GetContents(ctx, c.owner, repoName, filePath, nil)
	if err == nil && existing != nil {
		sha := existing.GetSHA()
		fileSHA = &sha
	}

	opts := &gogithub.RepositoryContentFileOptions{
		Message: gogithub.String(commitMsg),
		Content: []byte(base64.StdEncoding.EncodeToString(content)),
		SHA:     fileSHA,
	}

	if fileSHA == nil {
		_, _, err = c.gh.Repositories.CreateFile(ctx, c.owner, repoName, filePath, opts)
	} else {
		_, _, err = c.gh.Repositories.UpdateFile(ctx, c.owner, repoName, filePath, opts)
	}

	return err
}

// UploadDir uploads all files from a local directory to the repository,
// preserving relative path structure. Skips the .devcontainer subdirectory
// root (handled separately) and uploads devcontainer.json under .devcontainer/.
func (c *Client) UploadDir(ctx context.Context, repoName, localDir, commitPrefix string) error {
	return filepath.Walk(localDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(localDir, path)
		if err != nil {
			return err
		}

		// Use forward slashes for GitHub paths
		remotePath := filepath.ToSlash(rel)

		msg := fmt.Sprintf("%s: upload %s", commitPrefix, remotePath)
		return c.UploadFile(ctx, repoName, remotePath, path, msg)
	})
}
