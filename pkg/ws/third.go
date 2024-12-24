package ws

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/wuqunyong/file_storage/pkg/storage/controller"
	"github.com/wuqunyong/file_storage/pkg/storage/minio"
)

type ThirdApi struct {
	ws LongConnServer
}

func NewThirdApi(ws LongConnServer) *ThirdApi {
	return &ThirdApi{
		ws: ws,
	}
}

func (o *ThirdApi) InitiateMultipartUpload(c *gin.Context) {

	var conf minio.Config
	conf.Endpoint = "http://192.168.56.102:10005"
	conf.Bucket = "openimtest"
	conf.AccessKeyID = "root"
	conf.SecretAccessKey = "openIM123"
	conf.PublicRead = true
	minioCli, err := minio.NewMinio(c, conf)
	if err != nil {
		log.Printf("err:%s", err)
	}

	url, _ := minioCli.InitiateUpload(c, "hello_mp4", 1024, time.Duration(3600)*time.Second)
	log.Printf("url:%s", url)

	c.JSON(http.StatusOK, "InitiateMultipartUpload OK")
}

func (o *ThirdApi) CompleteMultipartUpload(c *gin.Context) {
	var conf minio.Config
	conf.Endpoint = "http://192.168.56.102:10005"
	conf.Bucket = "openimtest"
	conf.AccessKeyID = "root"
	conf.SecretAccessKey = "openIM123"
	conf.PublicRead = true
	minioCli, err := minio.NewMinio(c, conf)
	if err != nil {
		log.Printf("err:%s", err)
	}

	url, _ := minioCli.CompleteUpload(c, "hello_mp4_1024.mp4", "openimtest/temp/hello_mp4_1024.presigned")
	log.Printf("url:%s", url)

	c.JSON(http.StatusOK, "CompleteMultipartUpload OK")
}

func (o *ThirdApi) AccessURL(c *gin.Context) {
	var conf minio.Config
	conf.Endpoint = "http://192.168.56.102:10005"
	conf.Bucket = "openimtest"
	conf.AccessKeyID = "root"
	conf.SecretAccessKey = "openIM123"
	conf.PublicRead = true
	minioCli, err := minio.NewMinio(c, conf)
	if err != nil {
		log.Printf("err:%s", err)
	}

	req := controller.AccessURLOpt{}
	if err := c.BindJSON(&req); err != nil {
		return
	}

	url, _ := minioCli.AccessURL(c, "hello_mp4_1024.mp4", time.Duration(3600)*time.Second, req)
	log.Printf("url:%s", url)

	c.JSON(http.StatusOK, "AccessURL OK")
}

func (o *ThirdApi) WSHandler(c *gin.Context) {
	connContext := newContext(c.Writer, c.Request)

	// Create a WebSocket long connection object
	wsLongConn := newGWebSocket(1, time.Duration(6)*time.Second, 1024*1024)
	if err := wsLongConn.GenerateLongConn(c.Writer, c.Request); err != nil {
		return
	} else {
		if err := wsLongConn.RespondWithSuccess(); err != nil {
			// If the success message is successfully sent, end further processing
			return
		}
	}

	client := o.ws.GetClient()
	client.ResetClient(connContext, wsLongConn, o.ws, "123", NewClientHandler(client, NewRegisterHandler()))

	// Register the client with the server and start message processing
	o.ws.Register(client)
	client.Launch()
}
