package controllers

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"figorate/database"
	"figorate/helpers"
	"figorate/models"

	"github.com/getbrevo/brevo-go/lib"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type AuthController struct {
	userCollection         *mongo.Collection
	verificationCollection *mongo.Collection
}

func NewAuthController() *AuthController {
	return &AuthController{
		userCollection:         database.GetDatabase().Collection("users"),
		verificationCollection: database.GetDatabase().Collection("email_verifications"),
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
		IsActive:  false,
	}

	_, err = ac.userCollection.InsertOne(context.Background(), user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User registration failed"})
		return
	}

	// Generate verification token
	verificationToken, err := helpers.GenerateVerificationToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token generation failed"})
		return
	}

	// Store verification token
	verificationEntry := models.EmailVerificationToken{
		UserID:    user.ID,
		Token:     verificationToken,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	_, err = ac.verificationCollection.InsertOne(context.Background(), verificationEntry)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Verification token storage failed"})
		return
	}

	// Send verification email
	err = ac.sendVerificationEmail(user.Email, verificationToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Verification email sending failed"})
		return
	}

	accessToken, refreshToken, err := helpers.GenerateTokens(user.ID.Hex())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token generation failed"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully and verification mail sent", "access_token": accessToken, "refresh_token": refreshToken})
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

func (ac *AuthController) sendVerificationEmail(email, token string) error {
	apiKey := os.Getenv("BREVO_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("BREVO_API_KEY not set in environment")
	}

	verificationLink := fmt.Sprintf("%s/verify-email?token=%s", os.Getenv("APP_FRONTEND_URL"), token)

	cfg := lib.NewConfiguration()
	cfg.AddDefaultHeader("api-key", apiKey)

	client := lib.NewAPIClient(cfg)
	ctx := context.Background()

	// Create sender info
	sender := lib.SendSmtpEmailSender{
		Name:  "Figorate",
		Email: "noreply@yourdomain.com",
	}

	// Create HTML content
	htmlContent := fmt.Sprintf(`
		<html>
		<body>
			<h2>Welcome to Your Figorate!</h2>
			<p>Please verify your email by clicking the button below:</p>
			<a href="%s" style="background-color:#4CAF50;color:white;padding:10px 20px;text-decoration:none;border-radius:5px;">Verify Email</a>
		</body>
		</html>
	`, verificationLink)

	// Create email object
	sendSmtpEmail := lib.SendSmtpEmail{
		Sender:     &sender,
		To:         []lib.SendSmtpEmailTo{{Email: email}},
		Subject:    "Email Verification",
		HtmlContent: htmlContent,
	}

    apiInstance := client.TransactionalEmailsApi

	// Send email
	_, response, err := apiInstance.SendTransacEmail(ctx,sendSmtpEmail)
	if err != nil {
		return fmt.Errorf("failed to send email: %v, response: %v", err, response)
	}

	return nil
}

func (ac *AuthController) VerifyEmail(c *gin.Context) {
	token := c.Query("token")

	// Find and validate token
	var verificationEntry models.EmailVerificationToken
	err := ac.verificationCollection.FindOne(context.Background(),
		bson.M{
			"token":      token,
			"expires_at": bson.M{"$gt": time.Now()},
		}).Decode(&verificationEntry)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired token"})
		return
	}

	// Update user to active
	_, err = ac.userCollection.UpdateOne(
		context.Background(),
		bson.M{"_id": verificationEntry.UserID},
		bson.M{"$set": bson.M{"is_active": true}},
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User activation failed"})
		return
	}

	// Delete the verification token
	ac.verificationCollection.DeleteOne(
		context.Background(),
		bson.M{"token": token},
	)

	c.JSON(http.StatusOK, gin.H{"message": "Email verified successfully. You can now log in."})
}
