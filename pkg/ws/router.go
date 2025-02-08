package ws

import (
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	docs "github.com/wuqunyong/file_storage/docs"
)

func newGinRouter(ws LongConnServer) *gin.Engine {
	gin.SetMode(gin.DebugMode)
	r := gin.New()

	r.Use(gin.Recovery())

	docs.SwaggerInfo.BasePath = "/api/v1"

	// Third service
	t := NewThirdApi(ws)
	objectGroup := r.Group("/object")
	objectGroup.POST("/initiate_multipart_upload", t.InitiateMultipartUpload)
	objectGroup.POST("/complete_multipart_upload", t.CompleteMultipartUpload)
	objectGroup.POST("/access_url", t.AccessURL)

	r.GET("/ws", t.WSHandler)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	return r
}
