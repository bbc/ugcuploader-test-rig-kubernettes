package main

import (
	"encoding/gob"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"time"

	log "github.com/sirupsen/logrus"

	"golang.org/x/sync/errgroup"

	"github.com/gin-gonic/gin"

	"github.com/gin-gonic/gin/binding"

	ugl "github.com/bbc/ugcuploader-test-rig-kubernettes/fileupload/internal/pkg/ugcupload"
)

func init() {
	log.SetFormatter(&log.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})

}

var (
	g errgroup.Group
)

func Upload(c *gin.Context) {
	fileUpload := new(FileUpload)
	fop := ugl.FileUploadOperations{Context: c}
	if err := c.ShouldBindWith(fileUpload, binding.Form); err != nil {
		return
	}

	fop.SaveFile(fmt.Sprintf("%s/%s", "/data", fileUpload.Name))
}

func JmeterProps(c *gin.Context) {
	fileUpload := new(FileUpload)
	fop := ugl.FileUploadOperations{Context: c}
	if err := c.ShouldBindWith(fileUpload, binding.Form); err != nil {
		return
	}
	jh := os.Getenv("JMETER_HOME")
	fop.SaveFile(fmt.Sprintf("%s/bin/jmeter.properties", jh))
}

func UserProps(c *gin.Context) {
	fileUpload := new(FileUpload)
	fop := ugl.FileUploadOperations{Context: c}
	if err := c.ShouldBindWith(fileUpload, binding.Form); err != nil {
		return
	}
	jh := os.Getenv("JMETER_HOME")
	fop.SaveFile(fmt.Sprintf("%s/bin/user.properties", jh))
}

//StartServer used to start jmeter server
func StartServer(c *gin.Context) {

	cmd := exec.Command("/start.sh")
	out, err := cmd.CombinedOutput()
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
	}
	c.String(http.StatusOK, string(out))
	return
}
func main() {

	server01 := &http.Server{
		Addr:         ":1007",
		Handler:      router01(),
		ReadTimeout:  15 * time.Minute,
		WriteTimeout: 15 * time.Minute,
		IdleTimeout:  15 * time.Minute,
	}

	g.Go(func() error {
		return server01.ListenAndServe()
	})
	if err := g.Wait(); err != nil {
		log.Fatal(err)
	}

}

type FileUpload struct {
	File string `json:"file" form:"file"`
	Name string `json:"name" form:"name"`
}

func router01() http.Handler {
	// Gin instance
	r := gin.Default()

	gob.Register(FileUpload{})

	r.POST("/data", Upload)
	r.POST("/jmeter-props", JmeterProps)
	r.POST("/user-propes", UserProps)
	r.GET("/start-server", StartServer)

	return r
}
