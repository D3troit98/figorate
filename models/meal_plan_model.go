package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DailyMeals struct {
    Breakfast string `bson:"breakfast" json:"breakfast"`
    Lunch     string `bson:"lunch" json:"lunch"`
    Dinner    string `bson:"dinner" json:"dinner"`
    Dessert   string `bson:"dessert" json:"dessert"`
}

type MonthlyMealPlan struct {
    ID        primitive.ObjectID         `bson:"_id,omitempty" json:"id"`
    UserID    primitive.ObjectID         `bson:"user_id" json:"user_id"`
    Month     int                        `bson:"month" json:"month"`
    Year      int                        `bson:"year" json:"year"`
    Days      map[int]DailyMeals        `bson:"days" json:"days"` // Key is day of month (1-31)
    CreatedAt time.Time                 `bson:"created_at" json:"created_at"`
    UpdatedAt time.Time                 `bson:"updated_at" json:"updated_at"`
}
