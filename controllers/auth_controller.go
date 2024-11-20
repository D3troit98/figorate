package controllers

import (
	"context"
	"net/http"
	"os"
	"time"

	"figorate/database"
	"figorate/helpers"
	"figorate/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type AuthController struct {
	userCollection *mongo.Collection
}

func NewAuthController() *AuthController {
	return &AuthController{
		userCollection: database.GetDatabase().Collection("users"),
	}
}

func (ac *AuthController) SignUp(c *gin.Context) {
	var signUpRequest models.SignUpRequest
	if err := c.ShouldBindJSON(&signUpRequest); err != nil {
		// Generate user-friendly error message
		errorMessage := helpers.GenerateValidationError(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": errorMessage})
		return
	}

	// Validate password
	if err := helpers.ValidatePassword(signUpRequest.Password, signUpRequest.Email); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
	// Check if user already exists
	existingUser := ac.userCollection.FindOne(context.Background(), bson.M{"email": signUpRequest.Email})
	if existingUser.Err() == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
		return
	}

	// Hash password
	hashedPassword, err := helpers.HashPassword(signUpRequest.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Password hashing failed"})
		return
	}

	user := models.User{
		ID:        primitive.NewObjectID(),
		Email:     signUpRequest.Email,
		Password:  string(hashedPassword),
		FirstName: signUpRequest.FirstName,
		LastName:  signUpRequest.LastName,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		IsActive:  true,
	}

	_, err = ac.userCollection.InsertOne(context.Background(), user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User registration failed"})
		return
	}

	accessToken, refreshToken, err := helpers.GenerateTokens(user.ID.Hex())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token generation failed"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully", "access_token": accessToken, "refresh_token": refreshToken})
}

func (ac *AuthController) SignIn(c *gin.Context) {
	var signInRequest models.SignInRequest
	if err := c.ShouldBindJSON(&signInRequest); err != nil {
		errorMessage := helpers.GenerateValidationError(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": errorMessage})
		return
	}

	var user models.User
	err := ac.userCollection.FindOne(context.Background(), bson.M{"email": signInRequest.Email}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	err = helpers.CheckPasswordHash(signInRequest.Password, user.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	accessToken, refreshToken, err := helpers.GenerateTokens(user.ID.Hex())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token generation failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

func (ac *AuthController) RefreshToken(c *gin.Context) {
	refreshToken := c.GetHeader("Refresh-Token")
	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_REFRESH_SECRET")), nil
	})

	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
		return
	}

	userID := claims["user_id"].(string)
	accessToken, newRefreshToken, err := helpers.GenerateTokens(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token regeneration failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": newRefreshToken,
	})
}
