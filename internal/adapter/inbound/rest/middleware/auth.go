// internal/adapter/inbound/rest/middleware/auth.go
package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/bete7512/bill-aggregator/pkg/util/logger"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

// JWT claim structure
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// AuthMiddleware is a middleware for authentication
type AuthMiddleware struct {
	jwtSecret  []byte
	logger     *logger.Logger
	expiration time.Duration
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(jwtSecret string, expiration time.Duration, logger *logger.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		jwtSecret:  []byte(jwtSecret),
		logger:     logger,
		expiration: expiration,
	}
}

// Authenticate verifies the JWT token and sets the user ID in the context
func (m *AuthMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			return
		}

		// Extract the token from the header
		// Format: "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
			return
		}

		token := parts[1]

		userID, email, err := m.validateToken(token)
		if err != nil {
			m.logger.Error("Auth failed", "error", err.Error())
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		// Set user ID and email in context
		c.Set("userID", userID)
		c.Set("email", email)

		// Continue to the next handler
		c.Next()
	}
}

// GenerateToken generates a new JWT token for a user
func (m *AuthMiddleware) GenerateToken(userID, email string) (string, error) {
	// Set expiration time
	expirationTime := time.Now().Add(m.expiration)

	// Create claims
	claims := &Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "bill-aggregator",
			Subject:   userID,
		},
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret key
	tokenString, err := token.SignedString(m.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// validateToken validates a JWT token and returns the user ID and email
func (m *AuthMiddleware) validateToken(tokenString string) (string, string, error) {
	// Parse the token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.jwtSecret, nil
	})

	if err != nil {
		return "", "", fmt.Errorf("failed to parse token: %w", err)
	}

	// Get the claims
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims.UserID, claims.Email, nil
	}

	return "", "", errors.New("invalid token")
}

// RequirePermission middleware to check if user has the required permission
func (m *AuthMiddleware) RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context
		userID, exists := c.Get("userID")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		// In a real application, this would check the user's permissions
		// For this example, we'll simply allow all authenticated users
		m.logger.Info("Permission check", "userID", userID, "permission", permission)

		// Continue to the next handler
		c.Next()
	}
}

// ExtractUserID extracts the user ID from the context
func ExtractUserID(ctx context.Context) (string, error) {
	// For Gin context
	if ginCtx, ok := ctx.(*gin.Context); ok {
		if userID, exists := ginCtx.Get("userID"); exists {
			return userID.(string), nil
		}
		return "", errors.New("user ID not found in context")
	}

	// For standard context
	if userID, ok := ctx.Value("userID").(string); ok {
		return userID, nil
	}

	return "", errors.New("user ID not found in context")
}
