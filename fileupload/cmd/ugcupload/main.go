package main

import (
	"encoding/gob"
	"fmt"
	"net/http"
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	"golang.org/x/sync/errgroup"

	"github.com/gin-gonic/gin"

	"github.com/gin-gonic/gin/binding"

	shellExec "github.com/bbc/ugcuploader-test-rig-kubernettes/fileupload/internal/pkg/exec"
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

//Upload used to save the data file
func Upload(c *gin.Context) {
	fileUpload := new(FileUpload)
	fop := ugl.FileUploadOperations{Context: c}
	if err := c.ShouldBindWith(fileUpload, binding.Form); err != nil {
		log.WithFields(log.Fields{
			"err": err.Error(),
		}).Error("Problems binding form")
		return
	}
	fop.SaveFile(fmt.Sprintf("%s/%s", "/data", fileUpload.Name))
}

//JmeterProps used to save the jmeter
func JmeterProps(c *gin.Context) {
	fileUpload := new(FileUpload)
	fop := ugl.FileUploadOperations{Context: c}
	if err := c.ShouldBindWith(fileUpload, binding.Form); err != nil {
		log.WithFields(log.Fields{
			"err": err.Error(),
		}).Error("Problems binding form")
		return
	}
	jh := os.Getenv("JMETER_HOME")
	fop.SaveFile(fmt.Sprintf("%s/bin/jmeter.properties", jh))
}

//UserProps user to save the user.properties
func UserProps(c *gin.Context) {
	fileUpload := new(FileUpload)
	fop := ugl.FileUploadOperations{Context: c}
	if err := c.ShouldBindWith(fileUpload, binding.Form); err != nil {
		log.WithFields(log.Fields{
			"err": err.Error(),
		}).Error("Problems binding form")
		return
	}
	jh := os.Getenv("JMETER_HOME")
	fop.SaveFile(fmt.Sprintf("%s/bin/user.properties", jh))
}

//StartJmeterServer NOTE: Had to do this because the bash script was hanging...
func startJmeterServer() {

	cmd := fmt.Sprintf("/start.sh")
	args := []string{"<", "/dev/null"}
	se := shellExec.Exec{}
	_, err := se.ExecuteCommand(cmd, args)
	if err != "" {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("unable to start the test")
	}
}

//StartServer used to start jmeter server
func StartServer(c *gin.Context) {

	go startJmeterServer()
	//Just giving jmeter server time to start
	time.Sleep(2 * time.Second)
	c.String(http.StatusOK, "start test")
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

//FileUpload This is the struct returned
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
