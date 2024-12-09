package controllers

import (
	"context"
	"figorate/database"
	"figorate/models"
	"figorate/services"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type MealController struct {
	mealCollection     *mongo.Collection
	mealPlanCollection *mongo.Collection
	userCollection     *mongo.Collection
}

func NewMealController() *MealController {
	return &MealController{
		mealCollection:     database.GetDatabase().Collection("meals"),
		mealPlanCollection: database.GetDatabase().Collection("meal_plans"),
		userCollection:     database.GetDatabase().Collection("users"),
	}
}

func (mc *MealController) AddMeal(c *gin.Context) {
	var meal models.Meal
	if err := c.ShouldBindJSON(&meal); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
	meal.CreatedAt = time.Now()
	result, err := mc.mealCollection.InsertOne(context.Background(), meal)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add meal"})
		return
	}

	meal.ID = result.InsertedID.(primitive.ObjectID)
	c.JSON(http.StatusCreated, meal)

}

func (mc *MealController) GenerateMealPlan(c *gin.Context) {
	userIDHex, _ := c.Get("user_id")
	userID, _ := primitive.ObjectIDFromHex(userIDHex.(string))

	var user models.User
	err := mc.userCollection.FindOne(context.Background(), bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	filter := bson.M{
		"tags": bson.M{
			"$in": []string{user.NutritionPreference},
		},
	}
	var meals []models.Meal
	cursor, err := mc.mealCollection.Find(context.Background(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch meals"})
		return
	}
	defer cursor.Close(context.Background())

	if err := cursor.All(context.Background(), &meals); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to  process meal"})
		return
	}

	// Initialize AI service
	aiService := services.NewAIService(os.Getenv("OPENAI_API_KEY"))

	now := time.Now()
	daysInMonth := time.Date(now.Year(), now.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()

	// Generate meal plan using AI
	mealPlanDays, err := aiService.GenerateMealPlan(c.Request.Context(), services.MealPlanRequest{
		UserPreference: user.NutritionPreference,
		AvailableMeals: meals,
		DaysToGenerate: daysInMonth,
	})
	fmt.Print("meal plan generated ",mealPlanDays)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to generate meal plan: %v", err)})
		return
	}

	// Create monthly meal plan
	monthlyPlan := models.MonthlyMealPlan{
		UserID:    userID,
		Month:     int(now.Month()),
		Year:      now.Year(),
		Days:      mealPlanDays,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Save the meal plan
	_, err = mc.mealPlanCollection.DeleteMany(context.Background(), bson.M{
		"user_id": userID,
		"month":   monthlyPlan.Month,
		"year":    monthlyPlan.Year,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear existing meal plan"})
		return
	}

	result, err := mc.mealPlanCollection.InsertOne(context.Background(), monthlyPlan)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save meal plan"})
		return
	}

	monthlyPlan.ID = result.InsertedID.(primitive.ObjectID)
	c.JSON(http.StatusCreated, monthlyPlan)
}

// GetDailyMealPlan fetches a user's meal plan for a specific day
func (mc *MealController) GetDailyMealPlan(c *gin.Context) {
	userIDHex, _ := c.Get("user_id")
	userID, _ := primitive.ObjectIDFromHex(userIDHex.(string))

	// Get day from request
	dayStr := c.Param("day")
	day, err := strconv.Atoi(dayStr)
	if err != nil || day < 1 || day > 31 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid day"})
		return
	}

	// Get current month and year
	now := time.Now()

	var mealPlan models.MonthlyMealPlan
	err = mc.mealPlanCollection.FindOne(context.Background(), bson.M{
		"user_id": userID,
		"month":   int(now.Month()),
		"year":    now.Year(),
	}).Decode(&mealPlan)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Meal plan not found"})
		return
	}

	dailyMeals, exists := mealPlan.Days[day]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "No meal plan for this day"})
		return
	}

	c.JSON(http.StatusOK, dailyMeals)
}

// RecalibrateMealPlan updates the meal plan based on new preferences or requirements
func (mc *MealController) RecalibrateMealPlan(c *gin.Context) {
	userIDHex, _ := c.Get("user_id")
	userID, _ := primitive.ObjectIDFromHex(userIDHex.(string))

	// Optional: Allow specific days to be recalibrated
	var recalibrationRequest struct {
		Days []int `json:"days"` // Optional: specific days to recalibrate
	}
	if err := c.ShouldBindJSON(&recalibrationRequest); err != nil {
		// If no body provided, recalibrate all remaining days
		recalibrationRequest.Days = nil
	}

	// Get current meal plan
	now := time.Now()
	var mealPlan models.MonthlyMealPlan
	err := mc.mealPlanCollection.FindOne(context.Background(), bson.M{
		"user_id": userID,
		"month":   int(now.Month()),
		"year":    now.Year(),
	}).Decode(&mealPlan)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Meal plan not found"})
		return
	}

	// Get updated user preferences
	var user models.User
	err = mc.userCollection.FindOne(context.Background(), bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Fetch meals matching updated preferences
	filter := bson.M{
		"tags": bson.M{
			"$in": []string{user.NutritionPreference},
		},
	}

	var meals []models.Meal
	cursor, err := mc.mealCollection.Find(context.Background(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch meals"})
		return
	}
	defer cursor.Close(context.Background())

	if err := cursor.All(context.Background(), &meals); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process meals"})
		return
	}

	mealsByCategory := make(map[string][]models.Meal)
	for _, meal := range meals {
		mealsByCategory[meal.Category] = append(mealsByCategory[meal.Category], meal)
	}

	// Recalibrate specific days or remaining days in the month
	if len(recalibrationRequest.Days) > 0 {
		// Recalibrate only specified days
		for _, day := range recalibrationRequest.Days {
			if day >= 1 && day <= 31 {
				mealPlan.Days[day] = generateDailyMeals(mealsByCategory)
			}
		}
	} else {
		// Recalibrate all days from today onwards
		currentDay := now.Day()
		daysInMonth := time.Date(now.Year(), now.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()

		for day := currentDay; day <= daysInMonth; day++ {
			mealPlan.Days[day] = generateDailyMeals(mealsByCategory)
		}
	}

	mealPlan.UpdatedAt = now

	// Update the meal plan in database
	_, err = mc.mealPlanCollection.UpdateOne(
		context.Background(),
		bson.M{"_id": mealPlan.ID},
		bson.M{"$set": mealPlan},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update meal plan"})
		return
	}

	c.JSON(http.StatusOK, mealPlan)
}

// Helper function to generate
func generateDailyMeals(mealsByCategory map[string][]models.Meal) models.DailyMeals {
	return models.DailyMeals{
		Breakfast: getRandomMeal(mealsByCategory["breakfast"]).Name,
		Lunch:     getRandomMeal(mealsByCategory["lunch"]).Name,
		Dinner:    getRandomMeal(mealsByCategory["dinner"]).Name,
		Dessert:   getRandomMeal(mealsByCategory["dessert"]).Name,
	}
}

func getRandomMeal(meals []models.Meal) models.Meal {
	if len(meals) == 0 {
		return models.Meal{Name: "No meal available"}
	}
	return meals[rand.Intn(len(meals))]
}
