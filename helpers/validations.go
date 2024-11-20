package helpers

import (
	"figorate/models"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
)

// GenerateValidationError combines validation errors into a single string
func GenerateValidationError(err error) string {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		var errorMessages []string
		for _, validationErr := range validationErrors {
			field := strings.ToLower(validationErr.Field())
			errorMessages = append(errorMessages, field+" is missing")
		}

		// Combine messages and append "and x more" if there are multiple errors
		if len(errorMessages) > 1 {
			return errorMessages[0] + " and " + strconv.Itoa(len(errorMessages)-1) + " more"
		}
		return errorMessages[0]
	}

	// Return a generic error message if the error isn't validation-related
	return "Invalid input data"
}

// Additional validation for onboarding input
func ValidateOnboardingInput(req models.OnboardingRequest) error {


	// Example validation for health goals
	validHealthGoals := map[string]bool{
		"weight_loss": true,
		"muscle_gain": true,
		"improve_fitness": true,
		"stress_management": true,
		"improve_nutrition": true,
	}

	for _, goal := range req.HealthGoals {
		if !validHealthGoals[goal] {
			return fmt.Errorf("invalid health goal: %s", goal)
		}
	}


	validMedicalConditions := map[string]bool{
		"hypertension": true,
		"diabetes": true,
		"high_cholesterol": true,
		"asthma": true,
		"none":true,
		"kidney_disease":true,
		"cardiovascular_disease":true,
	}

	for _, condition := range req.MedicalConditions {
		if !validMedicalConditions[condition] {
			return fmt.Errorf("invalid medical condition: %s", condition)
		}
	}

	return nil
}
