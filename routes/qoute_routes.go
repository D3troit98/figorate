package routes

import (
	"figorate/controllers"
	"figorate/middleware"

	"github.com/gin-gonic/gin"
)

func SetupQouteRoutes(r *gin.Engine) {
	qouteController := controllers.NewQouteController()




	protectedRoutes := r.Group("/")
	protectedRoutes.Use(middleware.JWTAuthMiddleware())
	{
		protectedRoutes.GET("/qoutes",qouteController.GetQoutebyID)
		protectedRoutes.GET("/qoutes/random",qouteController.GetRandomQoute)
		protectedRoutes.GET("/qoutes/:id", qouteController.GetQoute)
		protectedRoutes.POST("/qoutes", qouteController.CreateQoute)
	}
}
