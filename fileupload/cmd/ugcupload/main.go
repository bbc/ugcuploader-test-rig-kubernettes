package main

import (
	"encoding/gob"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	ugl "github.com/bbc/ugcuploader-test-rig-kubernettes/fileupload/internal/pkg/ugcupload"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	pss "github.com/mitchellh/go-ps"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
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

//IsRunning Used to determing if slave is running
func IsRunning(c *gin.Context) {

	processes, err := pss.Processes()
	if err != nil {
		log.WithFields(log.Fields{
			"err": err.Error(),
		}).Error("Prolems listing all processes")
		c.String(http.StatusBadRequest, "no")
		return
	}
	for _, process := range processes {
		if strings.Contains(process.Executable(), "ApacheJMeter") || strings.Contains(process.Executable(), "java") {
			log.WithFields(log.Fields{
				"executable": process.Executable(),
				"pid":        process.Pid(),
				"parent pid": process.Pid(),
			}).Info("Processes")
			c.String(http.StatusOK, "yes")
			return
		}
	}

	c.String(http.StatusBadRequest, "no")
	return
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

func start(cmd string, args ...string) (p *os.Process, err error) {
	var procAttr os.ProcAttr
	procAttr.Files = []*os.File{os.Stdin,
		os.Stdout, os.Stderr}
	p, err = os.StartProcess(cmd, args, &procAttr)
	return
}

//StartJmeterServer NOTE: Had to do this because the bash script was hanging...
func startJmeterServer(startUpload StartUpload) {

	cmd := fmt.Sprintf("/start.sh")
	args := []string{startUpload.Xmx, startUpload.Xms, startUpload.MaxMetaspaceSize}
	process, err := start(cmd, args...)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("unable to start the test")
	}

	log.WithFields(log.Fields{
		"Pid":    process.Pid,
		"Params": strings.Join(args, ","),
	}).Info("PID of jmter server")

	time.Sleep(2 * time.Second)
	cmd = fmt.Sprintf("java")
	args = []string{"-jar", "/fileupload/jolokia-jvm-1.6.2-agent.jar", "start", "/opt/apache-jmeter/bin/ApacheJMeter.jar"}
	_, err = start(cmd, args...)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("unable to start the test")
	}

	log.WithFields(log.Fields{
		"Pid": process.Pid,
	}).Info("PID of Jolokia")

}

//StartServer used to start jmeter server
func StartServer(c *gin.Context) {

	startUpload := new(StartUpload)
	if err := c.ShouldBindWith(startUpload, binding.Form); err != nil {
		log.WithFields(log.Fields{
			"err": err.Error(),
		}).Error("Problems binding form to start the upload")
		c.String(http.StatusBadRequest, "Unable to start the test - problems binding form")
		return
	}

	go startJmeterServer(*startUpload)
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

//StartUpload holds the values for configuring the slave
type StartUpload struct {
	Xmx              string `json:"xmx" form:"xmx"`
	Xms              string `json:"xms" form:"xms"`
	MaxMetaspaceSize string `json:"maxMetaspaceSize" form:"maxMetaspaceSize"`
}

func router01() http.Handler {
	// Gin instance
	r := gin.Default()

	gob.Register(FileUpload{})
	gob.Register(StartUpload{})

	r.POST("/data", Upload)
	r.POST("/jmeter-props", JmeterProps)
	r.POST("/user-propes", UserProps)
	r.POST("/start-server", StartServer)
	r.GET("/is-running", IsRunning)

	return r
}
