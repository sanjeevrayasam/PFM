package controllers

import (
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/sanjeevrayasam/pfm-go/models"
	"golang.org/x/crypto/bcrypt"
)

// generateTokens generates new access and refresh tokens
func generateTokens(userID uint) (string, string, error) {
	// Define your secret keys for access and refresh tokens
	var accessSecretKey = "your_access_secret_key"   // Replace with actual secret key
	var refreshSecretKey = "your_refresh_secret_key" // Replace with actual secret key

	// Create an access token
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Minute * 15).Unix(), // short expiry for access token
	})
	accessTokenString, err := accessToken.SignedString([]byte(accessSecretKey))
	if err != nil {
		return "", "", err
	}

	// Create a refresh token
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24 * 7).Unix(), // longer expiry for refresh token
	})
	refreshTokenString, err := refreshToken.SignedString([]byte(refreshSecretKey))
	if err != nil {
		return "", "", err
	}

	return accessTokenString, refreshTokenString, nil
}

type AuthRequest struct {
	UserName string `json:"username"`
	Password string `json:"password"`
}

// @Summary Register new user
// @Description Register a new user in the system
// @Tags authentication
// @Accept  json
// @Produce  json
// @Param user body AuthRequest true "User to register"
// @Success 200 {object} models.User
// @Router /register [post]
func Register(c *gin.Context) {
	var request AuthRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	user := models.User{Username: request.UserName, Password: string(hashedPassword)}
	result := models.DB.Create(&user)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Registration successful"})

}

// @Summary User login
// @Description Logs in a user and returns a JWT token
// @Tags authentication
// @Accept  json
// @Produce  json
// @Param credentials body AuthRequest true "Login Credentials"
// @Success 200 {object} map[string]string "JWT Token"
// @Failure 400 {object} map[string]string "Invalid username or password"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /login [post]
func LoginUser(c *gin.Context) {
	var request AuthRequest
	var user models.User
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result := models.DB.Where("username=?", request.UserName).First(&user)
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid username or password"})
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.Password)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid username or password"})
		return
	}
	newAccessToken, newRefreshToken, err := generateTokens(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"access_token": newAccessToken, "refresh_token": newRefreshToken})
}

// RefreshTokenRequest defines the structure for the refresh token request payload
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// @Summary Refresh Access Token
// @Description Refreshes the access token using a refresh token
// @Tags authentication
// @Accept  json
// @Produce  json
// @Param  request body RefreshTokenRequest true "Refresh Token"
// @Success 200 {object} map[string]string "New Access and Refresh Tokens"
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Router /refresh [post]
func RefreshToken(c *gin.Context) {
	var request RefreshTokenRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// Parse the refresh token
	token, err := jwt.Parse(request.RefreshToken, func(token *jwt.Token) (interface{}, error) {
		return []byte("your_access_secret_key"), nil
	})

	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse token claims"})
		return
	}
	userID, ok := claims["user_id"].(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user IDfrom token claims"})
		return
	}
	// Generate new tokens
	newAccessToken, newRefreshToken, err := generateTokens(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  newAccessToken,
		"refresh_token": newRefreshToken,
	})

}
