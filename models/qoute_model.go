package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)


type Qoute struct {
	ID primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Content string `bson:"content" json:"content"`
	CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time          `bson:"updated_at" json:"updated_at"`
}

type CreateQouteRequest struct {
	Content string  `json:"content" binding:"required"`
}
