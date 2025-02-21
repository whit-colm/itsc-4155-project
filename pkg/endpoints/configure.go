package endpoints

import "github.com/gin-gonic/gin"

// Configure all backend endpoints
func Configure(router *gin.Engine) {
	router.GET("/api/books", GetBooks)
	router.GET("/api/health", Health)
}
