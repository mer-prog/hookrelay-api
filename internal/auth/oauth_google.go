package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/mer-prog/hookrelay-api/internal/config"
)

func GoogleOAuthConfig(cfg *config.Config) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     cfg.GoogleClientID,
		ClientSecret: cfg.GoogleClientSecret,
		Endpoint:     google.Endpoint,
		RedirectURL:  cfg.FrontendURL + "/auth/google/callback",
		Scopes:       []string{"openid", "email", "profile"},
	}
}

func GetGoogleUserInfo(ctx context.Context, oauthCfg *oauth2.Config, token *oauth2.Token) (*OAuthUser, error) {
	client := oauthCfg.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, fmt.Errorf("GetGoogleUserInfo: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("GetGoogleUserInfo read body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GetGoogleUserInfo: status %d: %s", resp.StatusCode, body)
	}

	var result struct {
		ID      string `json:"id"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("GetGoogleUserInfo unmarshal: %w", err)
	}

	return &OAuthUser{
		Email:      result.Email,
		Name:       result.Name,
		AvatarURL:  result.Picture,
		Provider:   "google",
		ProviderID: result.ID,
	}, nil
}
