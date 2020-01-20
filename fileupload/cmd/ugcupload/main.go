package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
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

//NOTE: Had to do this because the bash script was hanging...
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

// Code below taken from here: https://github.com/kjk/go-cookbook/blob/master/advanced-exec/03-live-progress-and-capture-v2.go
// CapturingPassThroughWriter is a writer that remembers
// data written to it and passes it to w
type CapturingPassThroughWriter struct {
	buf bytes.Buffer
	w   io.Writer
}

// NewCapturingPassThroughWriter creates new CapturingPassThroughWriter
func NewCapturingPassThroughWriter(w io.Writer) *CapturingPassThroughWriter {
	return &CapturingPassThroughWriter{
		w: w,
	}
}

// Write writes data to the writer, returns number of bytes written and an error
func (w *CapturingPassThroughWriter) Write(d []byte) (int, error) {
	w.buf.Write(d)
	return w.w.Write(d)
}

// Bytes returns bytes written to the writer
func (w *CapturingPassThroughWriter) Bytes() []byte {
	return w.buf.Bytes()
}
