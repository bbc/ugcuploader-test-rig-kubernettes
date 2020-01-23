package main

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"

	"golang.org/x/sync/errgroup"

	"github.com/gin-gonic/gin"

	shellExec "github.com/bbc/ugcuploader-test-rig-kubernettes/jmeter-master/internal/pkg/exec"

	pss "github.com/mitchellh/go-ps"

	"github.com/gin-gonic/gin/binding"

	"os/exec"
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

//StartTestCMD structure holding the data required to start a test
type StartTestCMD struct {
	TestFile string `json:"testfile" form:"testfile"`
	Tenant   string `json:"tenant" form:"tenant"`
	Hosts    string `json:"hosts" form:"hosts"`
}

//StartTest used to start the jmeter tests
func StartTest(c *gin.Context) {

	var startTestCMD StartTestCMD
	if er := c.ShouldBindWith(&startTestCMD, binding.Form); er != nil {
		log.WithFields(log.Fields{
			"err": er.Error(),
		}).Error("Problems binding form")
		return
	}

	var jsonData []byte
	jsonData, _ = json.Marshal(startTestCMD)

	log.WithFields(log.Fields{
		"startTestCmd": string(jsonData),
	}).Info("Command Received")

	processes, err := pss.Processes()
	if err != nil {
		log.WithFields(log.Fields{
			"err": err.Error(),
		}).Error("Prolems listing all processes")
		res := Response{}
		res.Message = "Prolems listing all processes"
		res.Code = 401
		c.PureJSON(http.StatusOK, res)
		return
	}
	for _, process := range processes {
		if strings.EqualFold(process.Executable(), "java") || strings.EqualFold(process.Executable(), "jmeter") {
			log.WithFields(log.Fields{
				"executable": process.Executable(),
				"pid":        process.Pid(),
				"parent pid": process.PPid(),
			}).Info("Processes")
			res := Response{}
			res.Message = "Test Are Running You will need to stop first"
			res.Code = 402
			c.PureJSON(http.StatusOK, res)
			return
		}
	}

	t := time.Now()
	u2 := fmt.Sprintf("%s-%s", uuid.NewV4(), t.Format("20060102150405"))
	path := fmt.Sprintf("/home/jmeter/test/%s", u2)
	errMkdir := os.MkdirAll(path, os.ModePerm)
	if errMkdir != nil {
		log.WithFields(log.Fields{
			"err":  errMkdir.Error(),
			"path": path,
		}).Error("Problems creating the directory")
		res := Response{}
		res.Message = "Unable to create the test path directory"
		res.Code = 404
		c.PureJSON(http.StatusBadRequest, res)
		return
	}

	jmeterScript, err := c.FormFile("file")
	if err != nil {
		log.WithFields(log.Fields{
			"err": err.Error(),
		}).Error("Unable to get the jmeter script from the form")
		res := Response{}
		res.Message = "Unable to get the jmeter script from the form"
		res.Code = 403
		c.PureJSON(http.StatusBadRequest, res)
		return
	}

	if jmeterScript != nil {
		destFileName := fmt.Sprintf("%s/upload.jmx", path)
		c.SaveUploadedFile(jmeterScript, destFileName)
	}

	args := fmt.Sprintf(" %s/upload.jmx %s %s ", path, startTestCMD.Tenant, startTestCMD.Hosts)

	cmd := exec.Command("/usr/local/bin/start_test_controller.sh", args)
	out, errExec := cmd.CombinedOutput()

	if errExec != nil {
		log.WithFields(log.Fields{
			"err": errExec.Error(),
		}).Error("Problems executing the script that starts jmeter")
		res := Response{}
		res.Message = "Unable to get the jmeter script from the form"
		res.Code = 404
		c.PureJSON(http.StatusBadRequest, res)
		return

	}

	log.WithFields(log.Fields{
		"out": string(out),
	}).Info("Tests were correctly started")

	res := Response{}
	res.Message = string(out)
	res.Code = 200
	c.PureJSON(http.StatusOK, res)
	return

}

//StopTest used to stop the tests
func StopTest(c *gin.Context) {

	cmd := fmt.Sprintf("/opt/apache-jmeter/bin/stoptest.sh")

	args := []string{"<", "/dev/null"}
	se := shellExec.Exec{}
	_, err := se.ExecuteCommand(cmd, args)
	if err != "" {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("unable to start the test")
	}
	return
}

//Response The response that is sent back to the caller
type Response struct {
	Message string
	Code    int
}

func main() {

	gob.Register(StartTestCMD{})
	server01 := &http.Server{
		Addr:         ":1025",
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

func router01() http.Handler {
	// Gin instance
	r := gin.Default()
	r.NoRoute(func(c *gin.Context) {
		res := Response{}
		res.Message = "no route defined"
		c.PureJSON(http.StatusOK, res)
	})

	r.GET("/stop-test", StopTest)
	r.POST("/start-test", StartTest)

	return r
}
