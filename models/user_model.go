package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Email          string             `bson:"email" json:"email" binding:"required,email"`
	Password       string             `bson:"password" json:"-"`
	FirstName      string             `bson:"first_name" json:"first_name"`
	LastName       string             `bson:"last_name" json:"last_name"`
	AuthProvider   string             `bson:"auth_provider" json:"auth_provider"`
	ProfilePicture string             `bson:"profile_picture" json:"profile_picture"`
	IsActive       bool               `bson:"is_active" json:"is_active"`
	LastLogin      time.Time          `bson:"last_login" json:"last_login"`
	CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time          `bson:"updated_at" json:"updated_at"`

	// New Onboarding Fields
	Gender              string   `bson:"gender" json:"gender"`
	Birthdate           string   `bson:"birthdate" json:"birthdate"`
	HealthGoals         []string `bson:"health_goals" json:"health_goals"`
	MedicalConditions   []string `bson:"medical_conditions" json:"medical_conditions"`
	NutritionPreference string   `bson:"nutrition_preference" json:"nutrition_preference"`
}

type SignUpRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
}

type OnboardingRequest struct {
	Gender              string   `json:"gender" binding:"required,oneof=male female other"`
	Birthdate           string   `json:"birthdate" binding:"required,datetime=2006-01-02"`
	HealthGoals         []string `json:"health_goals" binding:"required"`
	MedicalConditions   []string `json:"medical_condition" binding:"required"`
	NutritionPreference string   `json:"nutrition_preference" binding:"required,oneof=vegetarian vegan pescatarian gluten_free dairy_free none"`
}

type SignInRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type ResetPasswordRequest struct {
	Email       string `json:"email" binding:"required,email"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}
