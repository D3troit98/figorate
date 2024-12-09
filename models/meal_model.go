package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Meal struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name      string             `bson:"name" json:"name"`
	Image     string             `bson:"image" json:"image"`
	Calories  int                `bson:"calories" json:"calories"`
	Preptime  int                `bson:"prep_time" json:"prep_time"` // in minutes
	Category  string             `bson:"category" json:"category"`   // breakfast, lunch, etc.
	Tags      []string           `bson:"tags" json:"tags"`           // vegetarian, low-fat, etc.
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}
