package ws

import (
	"github.com/gin-gonic/gin"
)

func newGinRouter(ws LongConnServer) *gin.Engine {
	gin.SetMode(gin.DebugMode)
	r := gin.New()

	r.Use(gin.Recovery())

	// Third service
	t := NewThirdApi(ws)
	objectGroup := r.Group("/object")
	objectGroup.POST("/initiate_multipart_upload", t.InitiateMultipartUpload)
	objectGroup.POST("/complete_multipart_upload", t.CompleteMultipartUpload)
	objectGroup.POST("/access_url", t.AccessURL)

	r.GET("/ws", t.WSHandler)

	return r
}
