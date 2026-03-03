package auth

// OAuthUser represents a user profile retrieved from an OAuth provider.
type OAuthUser struct {
	Email      string `json:"email"`
	Name       string `json:"name"`
	AvatarURL  string `json:"avatar_url"`
	Provider   string `json:"provider"`
	ProviderID string `json:"provider_id"`
}
