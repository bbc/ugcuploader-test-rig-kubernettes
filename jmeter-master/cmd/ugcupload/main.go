package main

import (
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	pss "github.com/mitchellh/go-ps"
	uuid "github.com/satori/go.uuid"
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

//StartTestCMD structure holding the data required to start a test
type StartTestCMD struct {
	TestFile string `json:"testfile" form:"testfile"`
	Tenant   string `json:"tenant" form:"tenant"`
	Hosts    string `json:"hosts" form:"hosts"`
}

func runTest(args []string) {
	executeCommand("/home/jmeter/bin/load_test.sh", args)
}

func waitForPorts(hosts []string) (notReady bool) {

	count := 0
	temp := hosts
	var ready []string
	var waiting []string
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	fmt.Println("1:star")
	for {
		select {
		case <-ctx.Done():
			fmt.Println("2:star")
			for _, host := range temp {
				timeout := time.Second
				conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, "50000"), timeout)
				if err != nil {
					log.WithFields(log.Fields{
						"err":   err.Error(),
						"slave": host,
					}).Error("Problems connecting to slave")
				}
				if conn != nil {
					defer conn.Close()
					fmt.Println(fmt.Sprintf("NO READY:%s", host))
					ready = append(ready, host)
				} else {
					fmt.Println(fmt.Sprintf("READY:%s", host))
					waiting = append(waiting, host)
				}
			}
			temp = waiting
			waiting = []string{}
			fmt.Println("2:star")

			if len(ready) == len(hosts) {
				fmt.Println("All ports are ready")
				notReady = false
				return
			}

			count = count + 1
			if count == 10 {
				log.WithFields(log.Fields{
					"Slaves": strings.Join(waiting, ","),
				}).Error("slaves ports are not ready")
				notReady = true
				return
			}
			cancel()
			ctx, cancel = context.WithTimeout(context.Background(), 1*time.Minute)
		}
	}

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

	notReady := waitForPorts(strings.Split(startTestCMD.Hosts, ","))
	if notReady == true {
		res := Response{}
		res.Message = "slave ports were not opened"
		res.Code = 400
		c.PureJSON(http.StatusBadRequest, res)
		return
	}

	res := Response{}
	res.Message = "test should have started"
	res.Code = 200
	c.PureJSON(http.StatusOK, res)

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

	args := []string{fmt.Sprintf("%s/upload.jmx", path), startTestCMD.Tenant, startTestCMD.Hosts}
	log.WithFields(log.Fields{
		"args": strings.Join(args, ","),
	}).Info("Arguments being sent to jmeter script")

	go runTest(args)

	res = Response{}
	res.Message = "test should have started"
	res.Code = 200
	c.PureJSON(http.StatusOK, res)
	return

}

//IsRunning use to determine if the tenant is running
func IsRunning(c *gin.Context) {
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
			res.Message = "Test Are Running"
			res.Code = 200
			c.PureJSON(http.StatusOK, res)
			return
		}
	}

	res := Response{}
	res.Message = "No Test Running"
	res.Code = 400
	c.PureJSON(http.StatusOK, res)
	return
}

func executeCommand(c string, args []string) {
	cmd := exec.Command(c, args...)
	_, errExec := cmd.CombinedOutput()
	if errExec != nil {
		log.WithFields(log.Fields{
			"err": errExec.Error(),
		}).Error("Problems executing the script that starts jmeter")

	}

}

//StopTest used to stop the tests
func StopTest(c *gin.Context) {
	executeCommand(fmt.Sprintf("/opt/apache-jmeter/bin/stoptest.sh"), []string{})
	resp := Response{}
	resp.Message = "Test stopped"
	resp.Code = 200
	c.PureJSON(http.StatusOK, resp)
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
	r.GET("/is-running", IsRunning)

	return r
}
