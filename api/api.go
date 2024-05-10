package api

import (
	"qrcode/api/qrcode"
	"qrcode/middleware"

	"github.com/gin-gonic/gin"
)



func SetupAPI(router *gin.Engine) {
	apiGroup := router.Group("/api")
	apiGroup.Use(middleware.BearerTokenMiddleware())

	qrcodeApi := apiGroup.Group("/qrcode")
	{
		qrcodeApi.POST("/", qrcode.Create)
		qrcodeApi.GET("/", qrcode.GetQrcodes)
		qrcodeApi.GET("/:id", qrcode.GetQrcodeById)
	}	
}
