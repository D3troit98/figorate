package routes

import (
	"figorate/controllers"
	"figorate/middleware"

	"github.com/gin-gonic/gin"
)

func SetupAuthRoutes(r *gin.Engine) {
	authController := controllers.NewAuthController()
	userController := controllers.NewUserController()
	onboardingController := controllers.NewUserController()

	// Public routes
	r.POST("/signup", authController.SignUp)
	r.POST("/signin", authController.SignIn)
	r.POST("/refresh-token", authController.RefreshToken)

	// Protected routes
	protectedRoutes := r.Group("/")
	protectedRoutes.Use(middleware.JWTAuthMiddleware())
	{
		// Add protected routes here
		protectedRoutes.GET("/profile", userController.GetProfile)
		protectedRoutes.POST("/onboarding", onboardingController.CompleteOnboarding)
	}
}
