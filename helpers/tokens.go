package helpers

import (
	"crypto/rand"
	"encoding/base64"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// GenerateTokens generates access and refresh tokens for a user
func GenerateTokens(userID string) (string, string, error) {
	// Access token claims
	accessTokenClaims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 1).Unix(), // 1-hour expiration
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessTokenClaims)
	accessSecret := os.Getenv("JWT_SECRET")
	signedAccessToken, err := accessToken.SignedString([]byte(accessSecret))
	if err != nil {
		return "", "", err
	}

	// Refresh token claims
	refreshTokenClaims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24 * 7).Unix(), // 7-day expiration
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshTokenClaims)
	refreshSecret := os.Getenv("JWT_REFRESH_SECRET")
	signedRefreshToken, err := refreshToken.SignedString([]byte(refreshSecret))
	if err != nil {
		return "", "", err
	}

	return signedAccessToken, signedRefreshToken, nil
}

func GenerateVerificationToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
