package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"

	"github.com/mer-prog/hookrelay-api/internal/config"
)

func GitHubOAuthConfig(cfg *config.Config) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     cfg.GitHubClientID,
		ClientSecret: cfg.GitHubClientSecret,
		Endpoint:     github.Endpoint,
		RedirectURL:  cfg.FrontendURL + "/auth/github/callback",
		Scopes:       []string{"read:user", "user:email"},
	}
}

func GetGitHubUserInfo(ctx context.Context, oauthCfg *oauth2.Config, token *oauth2.Token) (*OAuthUser, error) {
	client := oauthCfg.Client(ctx, token)
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		return nil, fmt.Errorf("GetGitHubUserInfo: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("GetGitHubUserInfo read body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GetGitHubUserInfo: status %d: %s", resp.StatusCode, body)
	}

	var result struct {
		ID        int    `json:"id"`
		Login     string `json:"login"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("GetGitHubUserInfo unmarshal: %w", err)
	}

	name := result.Name
	if name == "" {
		name = result.Login
	}

	email := result.Email
	if email == "" {
		email, err = getGitHubPrimaryEmail(ctx, client)
		if err != nil {
			return nil, fmt.Errorf("GetGitHubUserInfo email: %w", err)
		}
	}

	return &OAuthUser{
		Email:      email,
		Name:       name,
		AvatarURL:  result.AvatarURL,
		Provider:   "github",
		ProviderID: strconv.Itoa(result.ID),
	}, nil
}

func getGitHubPrimaryEmail(ctx context.Context, client *http.Client) (string, error) {
	resp, err := client.Get("https://api.github.com/user/emails")
	if err != nil {
		return "", fmt.Errorf("fetching github emails: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading github emails: %w", err)
	}

	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}
	if err := json.Unmarshal(body, &emails); err != nil {
		return "", fmt.Errorf("unmarshaling github emails: %w", err)
	}

	for _, e := range emails {
		if e.Primary && e.Verified {
			return e.Email, nil
		}
	}
	return "", fmt.Errorf("no verified primary email found")
}
