package handler

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/oauth2"

	"github.com/mer-prog/hookrelay-api/internal/auth"
	"github.com/mer-prog/hookrelay-api/internal/config"
	"github.com/mer-prog/hookrelay-api/internal/model"
)

type AuthHandler struct {
	Pool       *pgxpool.Pool
	Config     *config.Config
	GoogleCfg  *oauth2.Config
	GitHubCfg  *oauth2.Config
}

// GoogleLogin handles GET /api/v1/auth/google
func (h *AuthHandler) GoogleLogin(c *gin.Context) {
	state := generateState()
	url := h.GoogleCfg.AuthCodeURL(state)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// GoogleCallback handles GET /api/v1/auth/google/callback
func (h *AuthHandler) GoogleCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing code"})
		return
	}

	token, err := h.GoogleCfg.Exchange(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "failed to exchange token"})
		return
	}

	oauthUser, err := auth.GetGoogleUserInfo(c.Request.Context(), h.GoogleCfg, token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user info"})
		return
	}

	h.handleOAuthUser(c, oauthUser)
}

// GitHubLogin handles GET /api/v1/auth/github
func (h *AuthHandler) GitHubLogin(c *gin.Context) {
	state := generateState()
	url := h.GitHubCfg.AuthCodeURL(state)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// GitHubCallback handles GET /api/v1/auth/github/callback
func (h *AuthHandler) GitHubCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing code"})
		return
	}

	token, err := h.GitHubCfg.Exchange(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "failed to exchange token"})
		return
	}

	oauthUser, err := auth.GetGitHubUserInfo(c.Request.Context(), h.GitHubCfg, token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user info"})
		return
	}

	h.handleOAuthUser(c, oauthUser)
}

// RefreshToken handles POST /api/v1/auth/refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	claims, err := auth.ValidateToken(h.Config.JWTSecret, req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}

	user, err := model.GetUserByID(c.Request.Context(), h.Pool, claims.GetUserUUID())
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}

	accessToken, err := auth.GenerateAccessToken(h.Config.JWTSecret, user.ID.String(), user.Email, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	refreshToken, err := auth.GenerateRefreshToken(h.Config.JWTSecret, user.ID.String())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate refresh token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

// Logout handles POST /api/v1/auth/logout (stateless — client deletes tokens)
func (h *AuthHandler) Logout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}

func (h *AuthHandler) handleOAuthUser(c *gin.Context, oauthUser *auth.OAuthUser) {
	ctx := c.Request.Context()

	user, err := model.GetUserByProvider(ctx, h.Pool, oauthUser.Provider, oauthUser.ProviderID)
	if err != nil {
		avatarURL := oauthUser.AvatarURL
		user = &model.User{
			Email:      oauthUser.Email,
			Name:       oauthUser.Name,
			AvatarURL:  &avatarURL,
			Provider:   oauthUser.Provider,
			ProviderID: oauthUser.ProviderID,
			Role:       "developer",
		}
		if err := model.CreateUser(ctx, h.Pool, user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
			return
		}
	}

	accessToken, err := auth.GenerateAccessToken(h.Config.JWTSecret, user.ID.String(), user.Email, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	refreshToken, err := auth.GenerateRefreshToken(h.Config.JWTSecret, user.ID.String())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate refresh token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"user":          user,
	})
}

func generateState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
