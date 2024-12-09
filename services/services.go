package services

import (
	"bytes"
	"context"
	"encoding/json"
	"figorate/models"
	"fmt"
	"net/http"
)




type AIService struct {
	apiKey string
	apiURL string
}

func NewAIService(apiKey string) *AIService{
	return &AIService{
		apiKey: apiKey,
		apiURL: "https://api.openai.com/v1/chat/completions",
	}
}

type MealPlanRequest struct {
	UserPreference string `json:"user_preference"`
	AvailableMeals []models.Meal `json:"available_meals"`
	DaysToGenerate int `json:"days_to_generate"`
}

type AIResponse struct {
	Choices []struct {
	Message struct {
		Content string `json:"content"`
	} `json:"message"`
	} `json:"choices"`
}

func (s *AIService) GenerateMealPlan(ctx context.Context, request MealPlanRequest) (map[int]models.DailyMeals, error){
	// Prepare the prompt for GPT
	prompt := fmt.Sprintf(`Given the following meals and user preference (%s), generate a balanced meal plan for %d days.
	Available meals: %v

	Rules:
	1. Only use meals from the provided list
	2. Ensure variety across days
	3. Match user's nutrition preference
	4. Balance caloric intake across meals
	5. Consider prep time distribution

	Return the meal plan as a JSON object with days as keys and meal names as values, following this structure:
	{
		"1": {"breakfast": "meal_name", "lunch": "meal_name", "dinner": "meal_name", "dessert": "meal_name"},
		...
	}`, request.UserPreference, request.DaysToGenerate, formatMealsForPrompt(request.AvailableMeals))

	// Prepare the API request
	reqBody := map[string]interface{}{
		"model": "gpt-4-turbo-preview",
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "You are a nutritionist and meal planning expert. Generate meal plans that are balanced and follow user preferences.",
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"temperature": 0.7,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", s.apiURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Parse response
	var aiResp AIResponse
	if err := json.NewDecoder(resp.Body).Decode(&aiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	// Parse the meal plan from the AI response
	var mealPlan map[int]models.DailyMeals
	if err := json.Unmarshal([]byte(aiResp.Choices[0].Message.Content), &mealPlan); err != nil {
		return nil, fmt.Errorf("failed to parse meal plan: %v", err)
	}

	return mealPlan, nil
}

func formatMealsForPrompt(meals []models.Meal) string{
	var result string
	for _,meal := range meals {
		result += fmt.Sprintf("-%s (Category: %s, Calories: %d, Tags: %v)\n",
		meal.Name, meal.Category,meal.Calories,meal.Tags)
	}
	return result
}
