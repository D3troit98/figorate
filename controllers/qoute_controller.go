package controllers

import (
	"context"
	"figorate/database"
	"figorate/models"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type QouteController struct {
	qouteCollection *mongo.Collection
	userCollection  *mongo.Collection
}

func NewQouteController() *QouteController {
	return &QouteController{
		qouteCollection: database.GetDatabase().Collection("qoutes"),
		userCollection:  database.GetDatabase().Collection("users"),
	}
}

func (qc *QouteController) CreateQoute(c *gin.Context) {
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
	err = qc.userCollection.FindOne(context.Background(), bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if user.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": " Only admins can create qoutes"})
		return
	}

	var request models.CreateQouteRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	qoute := models.Qoute{
		ID:        primitive.NewObjectID(),
		Content:   request.Content,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, err = qc.qouteCollection.InsertOne(context.Background(), qoute)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create qoute"})
		return
	}
	c.JSON(http.StatusCreated, qoute)
}

func (qc *QouteController) GetQoutebyID(c *gin.Context) {
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid qoute ID"})
		return
	}

	var qoute models.Qoute
	err = qc.qouteCollection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&qoute)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Qoute not found"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch qoute"})
		return
	}
	c.JSON(http.StatusOK, qoute)
}

func (qc *QouteController) GetQoute(c *gin.Context) {

	_, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	skip := (page - 1) * limit
	findOptions := options.Find()
	findOptions.SetSkip(int64(skip))
	findOptions.SetSkip(int64(limit))

	total, err := qc.qouteCollection.CountDocuments(context.Background(), bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count qoutes"})
		return
	}

	cursor, err := qc.qouteCollection.Find(context.Background(), bson.M{}, findOptions)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch qoutes"})
		return
	}
	defer cursor.Close(context.Background())

	var qoutes []models.Qoute
	if err := cursor.All(context.Background(), &qoutes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode qoutes"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"qoutes": qoutes,
		"pagination": gin.H{
			"current_page": page,
			"limit":        limit,
			"total":        total,
			"total_pages":  (total + int64(limit) - 1) / int64(limit),
		},
	})
}


func (qc *QouteController) GetRandomQoute(c * gin.Context){

	total, err :=qc.qouteCollection.CountDocuments(context.Background(),bson.M{})
	if err !=nil {
		c.JSON(http.StatusNotFound, gin.H{"error":"Failed to count qoutes"})
		return
	}
	if total == 0{
		c.JSON(http.StatusNotFound, gin.H{"error":"No qoutes available"})
		return
	}

	source := rand.New(rand.NewSource(time.Now().UnixNano()))
	randomSkip := source.Int63n(total)

	var qoute models.Qoute
	err = qc.qouteCollection.FindOne(
		context.Background(),
		bson.M{},
		options.FindOne().SetSkip(randomSkip),
	).Decode(&qoute)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error":"Failed to fetch random qoute"})
		return
	}
	c.JSON(http.StatusOK,qoute)
}
