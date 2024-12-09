package routes

import (
	"figorate/controllers"
	"figorate/middleware"

	"github.com/gin-gonic/gin"
)



func SetupMealRoutes(r *gin.Engine){
	mealController := controllers.NewMealController()


	mealRoutes := r.Group("/meals")
	mealRoutes.Use(middleware.JWTAuthMiddleware())
	{
		mealRoutes.POST("/add", mealController.AddMeal)
		mealRoutes.POST("/generate-plan",mealController.GenerateMealPlan)
		mealRoutes.GET("/plan/:day", mealController.GetDailyMealPlan)
		mealRoutes.POST("/recalibrate",mealController.RecalibrateMealPlan)
	}
}
