package endpoints

import "github.com/gin-gonic/gin"

// Configure all backend endpoints
func Configure(router *gin.Engine) {
	router.GET("/albums", GetAlbums)
	router.GET("/api/health", Health)
}
