package controllers

import (
	"context"
	"log"
	"net/http"
	"time"

	"figorate/database"
	"figorate/helpers"
	"figorate/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserController struct {
	userCollection *mongo.Collection
}

func NewUserController() *UserController {
	return &UserController{
		userCollection: database.GetDatabase().Collection("users"),
	}
}

func (uc *UserController) GetProfile(c *gin.Context) {
	// Get user ID from the JWT middleware
	userIDHex, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDHex.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	var user models.User
	err = uc.userCollection.FindOne(context.Background(), bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Sanitize the response by omitting sensitive fields
	profile := gin.H{
		"id":        user.ID,
		"email":     user.Email,
		"firstName": user.FirstName,
		"lastName":  user.LastName,
		"createdAt": user.CreatedAt,
	}

	c.JSON(http.StatusOK, profile)
}

func (uc *UserController) CompleteOnboarding(c *gin.Context) {
	userIDHex, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	log.Print(userIDHex)
	userID, err := primitive.ObjectIDFromHex(userIDHex.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	var onboardingRequest models.OnboardingRequest
	if err := c.ShouldBindJSON(&onboardingRequest); err != nil {
		errorMessage := helpers.GenerateValidationError(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": errorMessage})
		return
	}

	if err := helpers.ValidateOnboardingInput(onboardingRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	filter := bson.M{"_id": userID}
	update := bson.M{
		"$set": bson.M{
			"gender":               onboardingRequest.Gender,
			"birthdate":            onboardingRequest.Birthdate,
			"health_goals":         onboardingRequest.HealthGoals,
			"medical_conditions":   onboardingRequest.MedicalConditions,
			"nutrition_preference": onboardingRequest.NutritionPreference,
			"updated_at":           time.Now(),
		},
	}

	_, err = uc.userCollection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update onboarding information"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Onboarding completed successfully"})
}
